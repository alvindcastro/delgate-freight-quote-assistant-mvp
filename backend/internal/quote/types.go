package quote

import "time"

type Location struct {
	City       string `json:"city"`
	Province   string `json:"province"`
	PostalCode string `json:"postalCode"`
}

type Request struct {
	CustomerName  string   `json:"customerName"`
	CustomerEmail string   `json:"customerEmail"`
	Origin        Location `json:"origin"`
	Destination   Location `json:"destination"`

	ShipmentType string `json:"shipmentType"`
	Pieces       int    `json:"pieces"`
	Pallets      int    `json:"pallets"`

	WeightLbs float64 `json:"weightLbs"`
	LengthIn  float64 `json:"lengthIn"`
	WidthIn   float64 `json:"widthIn"`
	HeightIn  float64 `json:"heightIn"`

	ServiceLevel string   `json:"serviceLevel"`
	Accessorials []string `json:"accessorials"`
	Notes        string   `json:"notes"`
}

type Breakdown struct {
	LaneBase          float64            `json:"laneBase"`
	WeightCharge      float64            `json:"weightCharge"`
	DimensionalCharge float64            `json:"dimensionalCharge"`
	PieceHandling     float64            `json:"pieceHandling"`
	ShipmentSurcharge float64            `json:"shipmentSurcharge"`
	ServiceAdjustment float64            `json:"serviceAdjustment"`
	AccessorialFees   float64            `json:"accessorialFees"`
	FuelSurcharge     float64            `json:"fuelSurcharge"`
	AccessorialDetail map[string]float64 `json:"accessorialDetail"`
}

type Response struct {
	QuoteID   string    `json:"quoteId"`
	CreatedAt time.Time `json:"createdAt"`
	Currency  string    `json:"currency"`

	Origin      Location `json:"origin"`
	Destination Location `json:"destination"`
	RouteLabel  string   `json:"routeLabel"`

	EstimatedLow  float64 `json:"estimatedLow"`
	EstimatedHigh float64 `json:"estimatedHigh"`
	Midpoint      float64 `json:"midpoint"`

	ActualWeightLbs      float64 `json:"actualWeightLbs"`
	DimensionalWeightLbs float64 `json:"dimensionalWeightLbs"`
	ChargeableWeightLbs  float64 `json:"chargeableWeightLbs"`

	Status        string   `json:"status"`
	Confidence    string   `json:"confidence"`
	MissingFields []string `json:"missingFields"`
	RiskFlags     []string `json:"riskFlags"`
	NextSteps     []string `json:"nextSteps"`

	Breakdown Breakdown `json:"breakdown"`

	AssistantSummary  string   `json:"assistantSummary"`
	CustomerMessage   string   `json:"customerMessage"`
	FollowUpQuestions []string `json:"followUpQuestions"`
	AIProvider        string   `json:"aiProvider"`

	Disclaimer string `json:"disclaimer"`
}

type CalculateOptions struct {
	QuoteID   string
	CreatedAt time.Time
}
