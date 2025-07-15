package models

type Supporter struct {
	ID            string               `json:"id"`
	Name          string               `json:"name"`
	Rarity        Rarity               `json:"rarity"`
	Description   string               `json:"description"`
	TrainingBonus map[TrainingType]int `json:"training_bonus"`
	SpecialEffect string               `json:"special_effect"`
	IsOwned       bool                 `json:"is_owned"`
}

type Rarity int

const (
	Common Rarity = iota
	Rare
	SuperRare
	UltraRare
)

func (r Rarity) String() string {
	switch r {
	case Common:
		return "★"
	case Rare:
		return "★★"
	case SuperRare:
		return "★★★"
	case UltraRare:
		return "★★★★"
	default:
		return "?"
	}
}

func (r Rarity) Color() string {
	switch r {
	case Common:
		return "#888888"
	case Rare:
		return "#4CAF50"
	case SuperRare:
		return "#2196F3"
	case UltraRare:
		return "#FF9800"
	default:
		return "#FFFFFF"
	}
}

func NewSupporter(name, description string, rarity Rarity, bonuses map[TrainingType]int) *Supporter {
	return &Supporter{
		ID:            generateID(),
		Name:          name,
		Rarity:        rarity,
		Description:   description,
		TrainingBonus: bonuses,
		IsOwned:       false,
	}
}
