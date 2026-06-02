package parser

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/alvindcastro/delgate-freight-quote-assistant/backend/internal/quote"
)

type ParseRequest struct {
	Text string `json:"text"`
}

type ParseResponse struct {
	Draft           quote.Request `json:"draft"`
	ExtractedFields []string      `json:"extractedFields"`
	MissingHints    []string      `json:"missingHints"`
	Warnings        []string      `json:"warnings"`
}

func Parse(text string) ParseResponse {
	clean := strings.TrimSpace(text)
	lower := strings.ToLower(clean)
	draft := quote.Request{
		ShipmentType: "ltl_pallet",
		ServiceLevel: "standard",
		Pieces:       1,
		Pallets:      1,
		Notes:        clean,
	}
	var extracted []string
	var warnings []string

	if origin, destination, ok := parseRoute(clean); ok {
		draft.Origin.City = title(origin)
		draft.Destination.City = title(destination)
		if prov := cityProvince(origin); prov != "" {
			draft.Origin.Province = prov
		}
		if prov := cityProvince(destination); prov != "" {
			draft.Destination.Province = prov
		}
		extracted = append(extracted, "origin", "destination")
	}

	if pallets, ok := parseNumber(`(?i)(\d+)\s*pallets?`, clean); ok {
		draft.Pallets = int(pallets)
		draft.Pieces = int(pallets)
		draft.ShipmentType = "ltl_pallet"
		extracted = append(extracted, "pallet count")
	}

	if pieces, ok := parseNumber(`(?i)(\d+)\s*(pieces|pcs|items)`, clean); ok {
		draft.Pieces = int(pieces)
		extracted = append(extracted, "piece count")
	}

	if length, width, height, ok := parseDimensions(clean); ok {
		draft.LengthIn = length
		draft.WidthIn = width
		draft.HeightIn = height
		extracted = append(extracted, "dimensions")
	}

	if weight, ok := parseWeight(clean, draft.Pallets); ok {
		draft.WeightLbs = weight
		extracted = append(extracted, "weight")
	}

	if strings.Contains(lower, "same day") || strings.Contains(lower, "same-day") || strings.Contains(lower, "sameday") {
		draft.ServiceLevel = "same_day"
		extracted = append(extracted, "service level")
	} else if strings.Contains(lower, "expedited") || strings.Contains(lower, "rush") || strings.Contains(lower, "urgent") {
		draft.ServiceLevel = "expedited"
		extracted = append(extracted, "service level")
	}

	if containsAny(lower, []string{"furniture", "sofa", "couch", "table", "chair"}) {
		draft.ShipmentType = "furniture"
		extracted = append(extracted, "shipment type")
	}
	if containsAny(lower, []string{"appliance", "fridge", "refrigerator", "washer", "dryer", "stove"}) {
		draft.ShipmentType = "appliance"
		extracted = append(extracted, "shipment type")
	}
	if containsAny(lower, []string{"bulky", "oversized", "big and bulky"}) {
		draft.ShipmentType = "bulky_item"
		extracted = append(extracted, "shipment type")
	}

	draft.Accessorials = parseAccessorials(lower)
	if len(draft.Accessorials) > 0 {
		extracted = append(extracted, "accessorials")
	}

	if containsAny(lower, []string{"hazmat", "dangerous goods", "refrigerated", "temperature controlled", "customs", "border"}) {
		warnings = append(warnings, "Request contains special-handling language; operations review recommended")
	}

	missing := missingHints(draft)

	return ParseResponse{
		Draft:           draft,
		ExtractedFields: unique(extracted),
		MissingHints:    missing,
		Warnings:        warnings,
	}
}

func parseRoute(text string) (string, string, bool) {
	patterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)from\s+([a-zA-Z .'-]+?)\s+to\s+([a-zA-Z .'-]+?)(?:[.,;\n]|\s+(?:with|for|and|needs|requiring|before|by)\b|$)`),
		regexp.MustCompile(`(?i)([a-zA-Z .'-]+?)\s*(?:->|→)\s*([a-zA-Z .'-]+?)(?:[.,;\n]|$)`),
	}
	for _, pattern := range patterns {
		matches := pattern.FindStringSubmatch(text)
		if len(matches) == 3 {
			origin := strings.TrimSpace(matches[1])
			destination := strings.TrimSpace(matches[2])
			if origin != "" && destination != "" {
				return origin, destination, true
			}
		}
	}
	return "", "", false
}

func parseDimensions(text string) (float64, float64, float64, bool) {
	pattern := regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*(?:in|inch|inches)?\s*[x×]\s*(\d+(?:\.\d+)?)\s*(?:in|inch|inches)?\s*[x×]\s*(\d+(?:\.\d+)?)`)
	matches := pattern.FindStringSubmatch(text)
	if len(matches) != 4 {
		return 0, 0, 0, false
	}
	l, _ := strconv.ParseFloat(matches[1], 64)
	w, _ := strconv.ParseFloat(matches[2], 64)
	h, _ := strconv.ParseFloat(matches[3], 64)
	return l, w, h, true
}

func parseWeight(text string, pallets int) (float64, bool) {
	lower := strings.ToLower(text)
	pattern := regexp.MustCompile(`(?i)(\d+(?:\.\d+)?)\s*(lbs?|pounds?)`)
	matches := pattern.FindStringSubmatch(lower)
	if len(matches) != 3 {
		return 0, false
	}
	weight, _ := strconv.ParseFloat(matches[1], 64)
	if strings.Contains(lower, "each") && pallets > 1 {
		return weight * float64(pallets), true
	}
	return weight, true
}

func parseNumber(pattern, text string) (float64, bool) {
	re := regexp.MustCompile(pattern)
	matches := re.FindStringSubmatch(text)
	if len(matches) < 2 {
		return 0, false
	}
	value, err := strconv.ParseFloat(matches[1], 64)
	return value, err == nil
}

func parseAccessorials(lower string) []string {
	accessorials := make([]string, 0)
	keywords := map[string][]string{
		"liftgate":             {"liftgate", "lift gate"},
		"residential":          {"residential", "home delivery", "house"},
		"inside_delivery":      {"inside delivery", "inside"},
		"appointment_required": {"appointment", "delivery window", "scheduled"},
		"limited_access":       {"limited access", "mall", "school", "hospital", "construction site"},
		"fragile":              {"fragile", "glass", "delicate"},
		"oversized":            {"oversized", "overlength", "big and bulky"},
		"weekend":              {"weekend", "saturday", "sunday"},
	}
	for key, terms := range keywords {
		if containsAny(lower, terms) {
			accessorials = append(accessorials, key)
		}
	}
	return unique(accessorials)
}

func missingHints(req quote.Request) []string {
	var missing []string
	if req.Origin.City == "" {
		missing = append(missing, "origin city")
	}
	if req.Destination.City == "" {
		missing = append(missing, "destination city")
	}
	if req.Origin.PostalCode == "" {
		missing = append(missing, "pickup postal code")
	}
	if req.Destination.PostalCode == "" {
		missing = append(missing, "delivery postal code")
	}
	if req.WeightLbs <= 0 {
		missing = append(missing, "weight")
	}
	if req.LengthIn <= 0 || req.WidthIn <= 0 || req.HeightIn <= 0 {
		missing = append(missing, "dimensions")
	}
	return missing
}

func cityProvince(city string) string {
	key := strings.ToLower(strings.TrimSpace(city))
	key = strings.Trim(key, ".,;:")
	values := map[string]string{
		"vancouver":   "BC",
		"burnaby":     "BC",
		"surrey":      "BC",
		"richmond":    "BC",
		"kelowna":     "BC",
		"victoria":    "BC",
		"calgary":     "AB",
		"edmonton":    "AB",
		"red deer":    "AB",
		"regina":      "SK",
		"saskatoon":   "SK",
		"winnipeg":    "MB",
		"toronto":     "ON",
		"ottawa":      "ON",
		"mississauga": "ON",
		"brampton":    "ON",
		"montreal":    "QC",
		"montréal":    "QC",
		"halifax":     "NS",
	}
	return values[key]
}

func title(value string) string {
	value = strings.TrimSpace(value)
	parts := strings.Fields(strings.ToLower(value))
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, " ")
}

func containsAny(value string, terms []string) bool {
	for _, term := range terms {
		if strings.Contains(value, term) {
			return true
		}
	}
	return false
}

func unique(values []string) []string {
	seen := map[string]bool{}
	out := make([]string, 0, len(values))
	for _, value := range values {
		if !seen[value] {
			seen[value] = true
			out = append(out, value)
		}
	}
	return out
}
