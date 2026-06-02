package quote

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

const (
	statusReady        = "Ready to Quote"
	statusNeedsReview  = "Needs Review"
	statusManualReview = "Manual Review Required"
)

func Calculate(req Request, opts CalculateOptions) Response {
	req = normalizeRequest(req)
	if opts.CreatedAt.IsZero() {
		opts.CreatedAt = time.Now().UTC()
	}
	if opts.QuoteID == "" {
		opts.QuoteID = fmt.Sprintf("Q-%s", opts.CreatedAt.Format("20060102150405"))
	}

	pieces := req.Pieces
	if pieces <= 0 {
		pieces = 1
	}

	dimensionalWeight := calculateDimensionalWeight(req, pieces)
	chargeableWeight := math.Max(req.WeightLbs, dimensionalWeight)
	if chargeableWeight <= 0 {
		chargeableWeight = 150 // safe demo default so the estimate can still be calculated
	}

	missing := missingFields(req)
	riskFlags := riskFlags(req, chargeableWeight)

	laneBase := laneBase(req.Origin.Province, req.Destination.Province)
	weightCharge := roundMoney(math.Max(chargeableWeight, 150) * 0.32)
	dimensionalCharge := roundMoney(math.Max(0, dimensionalWeight-req.WeightLbs) * 0.08)
	pieceHandling := roundMoney(math.Max(0, float64(pieces-1)) * 18)
	shipmentSurcharge := shipmentSurcharge(req.ShipmentType)
	accessorialFees, accessorialDetail := accessorialFees(req.Accessorials)

	preServiceSubtotal := laneBase + weightCharge + dimensionalCharge + pieceHandling + shipmentSurcharge
	serviceMultiplier := serviceMultiplier(req.ServiceLevel)
	serviceAdjustment := roundMoney(preServiceSubtotal * (serviceMultiplier - 1))

	fuelBase := preServiceSubtotal + serviceAdjustment + accessorialFees
	fuelSurcharge := roundMoney(fuelBase * 0.16)
	midpoint := roundMoney(fuelBase + fuelSurcharge)

	estimatedLow := roundToFive(midpoint * 0.90)
	estimatedHigh := roundToFive(midpoint * 1.15)

	status := quoteStatus(missing, riskFlags)
	confidence := quoteConfidence(missing, riskFlags)
	nextSteps := buildNextSteps(status, missing, riskFlags)

	return Response{
		QuoteID:              opts.QuoteID,
		CreatedAt:            opts.CreatedAt,
		Currency:             "CAD",
		Origin:               req.Origin,
		Destination:          req.Destination,
		RouteLabel:           routeLabel(req.Origin, req.Destination),
		EstimatedLow:         estimatedLow,
		EstimatedHigh:        estimatedHigh,
		Midpoint:             midpoint,
		ActualWeightLbs:      roundMoney(req.WeightLbs),
		DimensionalWeightLbs: roundMoney(dimensionalWeight),
		ChargeableWeightLbs:  roundMoney(chargeableWeight),
		Status:               status,
		Confidence:           confidence,
		MissingFields:        missing,
		RiskFlags:            riskFlags,
		NextSteps:            nextSteps,
		Breakdown: Breakdown{
			LaneBase:          roundMoney(laneBase),
			WeightCharge:      weightCharge,
			DimensionalCharge: dimensionalCharge,
			PieceHandling:     pieceHandling,
			ShipmentSurcharge: roundMoney(shipmentSurcharge),
			ServiceAdjustment: serviceAdjustment,
			AccessorialFees:   roundMoney(accessorialFees),
			FuelSurcharge:     fuelSurcharge,
			AccessorialDetail: accessorialDetail,
		},
		Disclaimer: "Demo estimate only. Not a binding freight quote. Final pricing should be reviewed against carrier rates, postal-code zones, service availability, and operational constraints.",
	}
}

func normalizeRequest(req Request) Request {
	req.Origin.Province = normalizeProvince(req.Origin.Province)
	req.Destination.Province = normalizeProvince(req.Destination.Province)
	req.Origin.City = strings.TrimSpace(req.Origin.City)
	req.Destination.City = strings.TrimSpace(req.Destination.City)
	req.Origin.PostalCode = strings.ToUpper(strings.TrimSpace(req.Origin.PostalCode))
	req.Destination.PostalCode = strings.ToUpper(strings.TrimSpace(req.Destination.PostalCode))
	req.ShipmentType = normalizeKey(req.ShipmentType)
	if req.ShipmentType == "" {
		req.ShipmentType = "ltl_pallet"
	}
	req.ServiceLevel = normalizeKey(req.ServiceLevel)
	if req.ServiceLevel == "" {
		req.ServiceLevel = "standard"
	}
	for i := range req.Accessorials {
		req.Accessorials[i] = normalizeKey(req.Accessorials[i])
	}
	return req
}

func calculateDimensionalWeight(req Request, pieces int) float64 {
	if req.LengthIn <= 0 || req.WidthIn <= 0 || req.HeightIn <= 0 {
		return 0
	}
	// 139 is a common dimensional-weight divisor used for many parcel/air contexts.
	// This MVP uses it as a transparent demo heuristic, not as a production freight tariff.
	return (req.LengthIn * req.WidthIn * req.HeightIn * float64(pieces)) / 139
}

func laneBase(originProvince, destinationProvince string) float64 {
	originProvince = normalizeProvince(originProvince)
	destinationProvince = normalizeProvince(destinationProvince)

	if originProvince == "" || destinationProvince == "" {
		return 280
	}
	if originProvince == destinationProvince {
		return 175
	}
	if isNorthern(originProvince) || isNorthern(destinationProvince) {
		return 725
	}

	originRegion := region(originProvince)
	destinationRegion := region(destinationProvince)
	if originRegion == destinationRegion {
		return 245
	}

	if adjacentRegions(originRegion, destinationRegion) {
		return 390
	}

	return 650
}

func shipmentSurcharge(shipmentType string) float64 {
	switch normalizeKey(shipmentType) {
	case "parcel":
		return 15
	case "furniture":
		return 75
	case "appliance":
		return 85
	case "bulky_item", "big_and_bulky", "oversized":
		return 95
	case "ltl_pallet", "pallet":
		return 30
	default:
		return 45
	}
}

func serviceMultiplier(serviceLevel string) float64 {
	switch normalizeKey(serviceLevel) {
	case "expedited", "rush":
		return 1.30
	case "same_day", "same-day", "sameday":
		return 1.85
	default:
		return 1.00
	}
}

func accessorialFees(accessorials []string) (float64, map[string]float64) {
	fees := map[string]float64{
		"liftgate":             75,
		"residential":          85,
		"inside_delivery":      95,
		"appointment_required": 35,
		"limited_access":       90,
		"fragile":              40,
		"oversized":            120,
		"weekend":              80,
	}

	detail := make(map[string]float64)
	total := 0.0
	for _, raw := range accessorials {
		key := normalizeKey(raw)
		if fee, ok := fees[key]; ok {
			if _, seen := detail[key]; !seen {
				detail[key] = fee
				total += fee
			}
		}
	}
	return total, detail
}

func missingFields(req Request) []string {
	var missing []string
	if req.Origin.City == "" {
		missing = append(missing, "Origin city")
	}
	if req.Origin.Province == "" {
		missing = append(missing, "Origin province")
	}
	if req.Origin.PostalCode == "" {
		missing = append(missing, "Pickup postal code")
	}
	if req.Destination.City == "" {
		missing = append(missing, "Destination city")
	}
	if req.Destination.Province == "" {
		missing = append(missing, "Destination province")
	}
	if req.Destination.PostalCode == "" {
		missing = append(missing, "Delivery postal code")
	}
	if req.WeightLbs <= 0 {
		missing = append(missing, "Shipment weight")
	}
	if req.LengthIn <= 0 || req.WidthIn <= 0 || req.HeightIn <= 0 {
		missing = append(missing, "Shipment dimensions")
	}
	if req.Pieces <= 0 && req.Pallets <= 0 {
		missing = append(missing, "Piece or pallet count")
	}
	return missing
}

func riskFlags(req Request, chargeableWeight float64) []string {
	var risks []string
	if req.WeightLbs > 2500 || chargeableWeight > 3500 {
		risks = append(risks, "Heavy shipment should be manually reviewed")
	}
	if req.LengthIn > 96 || req.WidthIn > 60 || req.HeightIn > 96 {
		risks = append(risks, "Oversized dimensions may require special handling")
	}
	if normalizeKey(req.ServiceLevel) == "same_day" && normalizeProvince(req.Origin.Province) != normalizeProvince(req.Destination.Province) {
		risks = append(risks, "Same-day cross-province delivery requires operations review")
	}

	notes := strings.ToLower(req.Notes)
	watchWords := []string{"hazmat", "dangerous", "temperature", "refrigerated", "customs", "border", "white glove", "assembly"}
	for _, word := range watchWords {
		if strings.Contains(notes, word) {
			risks = append(risks, fmt.Sprintf("Notes mention %q; verify requirements", word))
		}
	}

	sort.Strings(risks)
	return uniqueStrings(risks)
}

func quoteStatus(missing, risks []string) string {
	for _, r := range risks {
		if strings.Contains(strings.ToLower(r), "heavy") || strings.Contains(strings.ToLower(r), "hazmat") || strings.Contains(strings.ToLower(r), "same-day") {
			return statusManualReview
		}
	}
	if len(missing) > 0 || len(risks) > 0 {
		return statusNeedsReview
	}
	return statusReady
}

func quoteConfidence(missing, risks []string) string {
	if len(missing) == 0 && len(risks) == 0 {
		return "High"
	}
	if len(missing) <= 2 && len(risks) <= 1 {
		return "Medium"
	}
	return "Low"
}

func buildNextSteps(status string, missing, risks []string) []string {
	var steps []string
	if len(missing) > 0 {
		steps = append(steps, "Collect missing shipment details before sending a final quote")
	}
	if len(risks) > 0 {
		steps = append(steps, "Send flagged items to operations for review")
	}
	switch status {
	case statusReady:
		steps = append(steps, "Quote is ready for customer review")
	case statusNeedsReview:
		steps = append(steps, "Confirm details and then convert the estimate into a formal quote")
	case statusManualReview:
		steps = append(steps, "Do not send as final pricing until operations approves the shipment")
	}
	return uniqueStrings(steps)
}

func routeLabel(origin, destination Location) string {
	left := strings.TrimSpace(strings.Join(nonEmpty([]string{origin.City, origin.Province}), ", "))
	right := strings.TrimSpace(strings.Join(nonEmpty([]string{destination.City, destination.Province}), ", "))
	if left == "" {
		left = "Unknown origin"
	}
	if right == "" {
		right = "Unknown destination"
	}
	return left + " → " + right
}

func normalizeProvince(value string) string {
	v := strings.ToUpper(strings.TrimSpace(value))
	v = strings.TrimSuffix(v, ".")
	provinces := map[string]string{
		"BRITISH COLUMBIA":          "BC",
		"ALBERTA":                   "AB",
		"SASKATCHEWAN":              "SK",
		"MANITOBA":                  "MB",
		"ONTARIO":                   "ON",
		"QUEBEC":                    "QC",
		"QUÉBEC":                    "QC",
		"NEW BRUNSWICK":             "NB",
		"NOVA SCOTIA":               "NS",
		"PRINCE EDWARD ISLAND":      "PE",
		"PEI":                       "PE",
		"NEWFOUNDLAND":              "NL",
		"NEWFOUNDLAND AND LABRADOR": "NL",
		"YUKON":                     "YT",
		"NORTHWEST TERRITORIES":     "NT",
		"NUNAVUT":                   "NU",
	}
	if replacement, ok := provinces[v]; ok {
		return replacement
	}
	return v
}

func normalizeKey(value string) string {
	v := strings.ToLower(strings.TrimSpace(value))
	v = strings.ReplaceAll(v, "-", "_")
	v = strings.ReplaceAll(v, " ", "_")
	return v
}

func region(province string) string {
	switch normalizeProvince(province) {
	case "BC", "AB", "SK", "MB":
		return "west"
	case "ON", "QC":
		return "central"
	case "NB", "NS", "PE", "NL":
		return "atlantic"
	case "YT", "NT", "NU":
		return "north"
	default:
		return "unknown"
	}
}

func adjacentRegions(a, b string) bool {
	if a == b {
		return true
	}
	pair := map[string]bool{
		"west:central":     true,
		"central:west":     true,
		"central:atlantic": true,
		"atlantic:central": true,
	}
	return pair[a+":"+b]
}

func isNorthern(province string) bool {
	r := region(province)
	return r == "north"
}

func roundMoney(v float64) float64 {
	return math.Round(v*100) / 100
}

func roundToFive(v float64) float64 {
	return math.Round(v/5) * 5
}

func nonEmpty(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			out = append(out, strings.TrimSpace(value))
		}
	}
	return out
}

func uniqueStrings(values []string) []string {
	seen := make(map[string]bool)
	out := make([]string, 0, len(values))
	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	return out
}
