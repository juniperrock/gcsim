package celestialgift

import (
	"fmt"

	"github.com/genshinsim/gcsim/pkg/core"
	"github.com/genshinsim/gcsim/pkg/core/attributes"
	"github.com/genshinsim/gcsim/pkg/core/event"
	"github.com/genshinsim/gcsim/pkg/core/info"
	"github.com/genshinsim/gcsim/pkg/core/keys"
	"github.com/genshinsim/gcsim/pkg/core/player/character"
	"github.com/genshinsim/gcsim/pkg/modifier"
)

const (
	duration      = 60 * 20
	set4Key       = "celestialgift-4pc"
	set4KeyHolder = set4Key + "-holder"
	set4KeyActive = set4Key + "-active"
)

func init() {
	core.RegisterSetFunc(keys.CelestialGift, NewSet)
}

type Set struct {
	Index int
	Count int
	buff  []float64
	core  *core.Core
	char  *character.CharWrapper
}

func (s *Set) SetIndex(idx int) { s.Index = idx }
func (s *Set) GetCount() int    { return s.Count }
func (s *Set) Init() error      { return nil }

func NewSet(c *core.Core, char *character.CharWrapper, count int, param map[string]int) (info.Set, error) {
	s := Set{
		Count: count,
		buff:  make([]float64, attributes.EndStatType),
		core:  c,
		char:  char,
	}

	if count < 2 {
		return &s, nil
	}

	m := make([]float64, attributes.EndStatType)
	m[attributes.ER] = 0.2
	char.AddStatMod(character.StatMod{
		Base:         modifier.NewBase("celestialgift-2pc", -1),
		AffectedStat: attributes.ER,
		Amount: func() []float64 {
			return m
		},
	})

	if count < 4 || !char.IsHexerei {
		return &s, nil
	}

	holderBuffActive := func() bool {
		for _, c := range c.Player.Chars() {
			if c.StatusIsActive(set4KeyHolder) {
				return true
			}
		}
		return false
	}

	holderBuffDuration := func() int {
		x := 0
		for _, c := range c.Player.Chars() {
			x = max(x, c.StatusDuration(set4KeyHolder))
		}
		return x
	}

	holderBuffActiveOfElement := func(e attributes.Element) bool {
		for _, c := range c.Player.Chars() {
			if c.StatusIsActive(set4KeyHolder) && c.Base.Element == e {
				return true
			}
		}
		return false
	}

	holderBuffActiveOfElementNotFrom := func(e attributes.Element, i int) bool {
		for _, c := range c.Player.Chars() {
			if c.StatusIsActive(set4KeyHolder) && c.Base.Element == e && c.Index() != i {
				return true
			}
		}
		return false
	}

	activeBuffActive := func() bool {
		for _, c := range c.Player.Chars() {
			if c.StatusIsActive(set4KeyActive) {
				return true
			}
		}
		return false
	}

	clearActiveBuff := func() {
		for _, x := range c.Player.Chars() {
			x.DeleteStatus(set4KeyActive)
		}
	}

	elems := map[attributes.Element]bool{}
	for _, x := range c.Player.Chars() {
		elems[x.Base.Element] = true
	}

	buffStrength := 0.2
	isHexereiSecretRite := c.Player.GetHexereiCount() >= 2
	if isHexereiSecretRite {
		buffStrength = 0.4
	}

	c.Events.Subscribe(event.OnSkill, func(args ...any) {
		// NOTE: Description says "The equipping character can trigger this effect while off-field."
		// but currently no characters can use skills while off field, so this is ignored
		if c.Player.Active() != char.Index() {
			return
		}
		clearActiveBuff()
		elem := char.Base.Element
		if holderBuffActiveOfElementNotFrom(elem, char.Index()) {
			return
		}
		char.AddStatus(set4KeyHolder, duration, true)
		buffedDmgP := attributes.EleToDmgP(elem)
		for _, x := range c.Player.Chars() {
			x.AddStatMod(character.StatMod{
				Base:         modifier.NewBase(set4KeyHolder+"-"+elem.String(), -1),
				AffectedStat: buffedDmgP,
				Amount: func() []float64 {
					clear(s.buff)
					s.buff[buffedDmgP] = buffStrength
					if holderBuffActiveOfElement(elem) {
						return s.buff
					}
					return nil
				},
			})
		}
	}, fmt.Sprintf(set4KeyHolder+"-%v", char.Base.Key.String()))

	if !isHexereiSecretRite {
		return &s, nil
	}

	c.Events.Subscribe(event.OnCharacterSwap, func(args ...any) {
		clearActiveBuff()
		remainingDuration := holderBuffDuration()
		if remainingDuration == 0 {
			return
		}
		elem := c.Player.ActiveChar().Base.Element
		if holderBuffActiveOfElement(elem) {
			return
		}
		char.AddStatus(set4KeyActive, remainingDuration, true)
		buffedDmgP := attributes.EleToDmgP(elem)
		for _, x := range c.Player.Chars() {
			x.AddStatMod(character.StatMod{
				Base:         modifier.NewBase(set4KeyHolder+"-"+elem.String()+"-active", -1),
				AffectedStat: buffedDmgP,
				Amount: func() []float64 {
					clear(s.buff)
					s.buff[buffedDmgP] = buffStrength
					if activeBuffActive() && holderBuffActive() {
						return s.buff
					}
					return nil
				},
			})
		}
	}, fmt.Sprintf(set4KeyActive+"-%v", char.Base.Key.String()))

	return &s, nil
}
