package displate

const (
	StatusAvailable = "active"
	StatusSoldOut   = "sold_out"
	StatusUpcoming  = "upcoming"
)

type LimitedEditionResponse struct {
	Data []Displate `json:"data"`
}
type Image struct {
	URL string `json:"url"`
	Alt any    `json:"alt"`
}
type Images struct {
	Main Image `json:"main"`
}
type Edition struct {
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	Status      string `json:"status"`
	Available   int    `json:"available"`
	Size        int    `json:"size"`
	Type        string `json:"type"`
	Format      string `json:"format"`
	TimeToStart int    `json:"timeToStart"`
}
type Displate struct {
	ID               int     `json:"id"`
	ItemCollectionID int     `json:"itemCollectionId"`
	Title            string  `json:"title"`
	URL              string  `json:"url"`
	Edition          Edition `json:"edition,omitempty"`
	Images           Images  `json:"images"`
}

func FilterDisplates(displates []Displate, filterFunc func(displate Displate) bool) []Displate {
	filtered := make([]Displate, 0, len(displates))
	for _, d := range displates {
		if filterFunc(d) {
			filtered = append(filtered, d)
		}
	}

	return filtered
}
