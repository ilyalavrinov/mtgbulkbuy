package mtgbulk

type NamesRequest struct {
	Cards map[string]int
}

func NewNamesRequest() NamesRequest {
	return NamesRequest{
		Cards: make(map[string]int),
	}
}

type PlatformType int

const (
	MtgSale  PlatformType = 0
	MtgTrade PlatformType = 1
)

type CardPrice struct {
	Price    int
	Platform PlatformType
	Shop     string
	URL      string
}

type CardResult struct {
	Found  bool
	Prices []CardPrice
}

type NamesResult struct {
	Cards map[string]CardResult
}

func ProcessByNames(cards NamesRequest) (*NamesResult, error) {
	logger.Debugw("Incoming ProcessByNames request",
		"count", len(cards.Cards))
	return nil, nil
}
