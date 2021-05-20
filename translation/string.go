package translation

import (
	"encoding/json"
	"strings"
)

type String struct {
	Display   string            `json:"display" yaml:"display"`
	Second    string            `json:"second" yaml:"second"`
	Translate map[string]string `json:"translate" yaml:"translate"`

	// properties which helping prevent repeated applying translation context
	ctxApplied bool
}

func NewString(lang, str string) String {
	return String{
		Translate: map[string]string{
			lang: str,
		},
	}
}

func (o *String) Init() *String {
	if o.Translate == nil {
		o.Translate = make(map[string]string)
	}

	return o
}

func (o *String) Reset() {
	o.Display = ""
	o.Second = ""
	o.Translate = nil
	o.ctxApplied = false
}

// A response with translation string contains ctxApplied as false.
// Reset the ctxApplied will be useful in test to compare a response and an expected value.
func (o *String) ResetCtxApplied() {
	o.ctxApplied = false
}

func (o *String) resetTranslation() *String {
	o.Display = ""
	o.Second = ""
	o.ctxApplied = false

	return o
}

func (o *String) ClearContext() *String {
	o.Display = ""
	o.Second = ""

	return o
}

func (o *String) ApplyTranslationCtx(ctx Context) *String {
	if o == nil || o.ctxApplied {
		return o
	}

	o.ctxApplied = true

	if str, ok := o.Translate[ctx.GetDisplay()]; ok && str != "" {
		o.Display = str
	} else {
		o.Display = o.Translate[ctx.GetFallback()]
	}

	o.Second = o.Translate[ctx.GetSecond()]

	if !ctx.GetTranslationList() {
		o.Translate = nil
	}

	return o
}

func (o *String) Clone() *String {
	if o == nil {
		return nil
	}

	cloned := *o

	if len(o.Translate) > 0 {
		cloned.Translate = make(map[string]string)

		for lang, translation := range o.Translate {
			cloned.Translate[lang] = translation
		}
	}

	return &cloned
}

// Checking an empty in a source of translation
func (o *String) Empty() bool {
	for _, str := range o.Translate {
		if len(str) > 0 {
			return false
		}
	}

	return true
}

// Checking a whole state of a string - a source of translation and a result of it
func (o *String) HasTranslation() bool {
	return !o.Empty() || o.Display != "" || o.Second != ""
}

func (o *String) Update(r String) {
	o.Init()

	for lang, str := range r.Translate {
		o.Translate[strings.ToLower(lang)] = str
	}
}

func (o *String) Add(r String) {
	o.Init()

	for lang, str := range r.Translate {
		o.Translate[strings.ToLower(lang)] += str
	}
}

func (o *String) AddTranslate(lang, str string) *String {
	o.Init()

	o.Translate[lang] = str

	return o
}

func (o *String) Map(f func(string) string) String {
	if len(o.Translate) == 0 {
		return String{}
	}

	s := (&String{}).Init()

	for lang, str := range o.Translate {
		s.Translate[strings.ToLower(lang)] = f(str)
	}

	return *s
}

func (o *String) Join(r String, s string) String {
	joined := *o.Clone().resetTranslation()

	for lang, str := range r.Translate {
		lang = strings.ToLower(lang)

		if _, ok := joined.Translate[lang]; !ok {
			joined.Translate[lang] = str
			continue
		}

		joined.Translate[lang] += s + str
	}

	return joined
}

func (o *String) Trim() int {
	*o = o.Map(func(s string) string { return strings.TrimSpace(s) })

	return o.Len()
}

func (o *String) Len() int {
	ln := 0

	for _, str := range o.Translate {
		if len(str) > ln {
			ln = len(str)
		}
	}

	return ln
}

func (o *String) GetTranslate(lang string) string {
	return o.Translate[lang]
}

func (o *String) String() string {
	data, _ := json.Marshal(o.Translate)

	return string(data)
}

func (o *String) UnmarshalJSON(data []byte) error {
	type Alias String
	aux := struct {
		*Alias
		TranslateBuf struct{
		En string `json:"en"`
		Ru string `json:"ru"`
	} `json:"translate"`
	}{
		Alias: (*Alias)(o),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	if o.Translate == nil {
		o.Translate = map[string]string{}
	}
	if aux.TranslateBuf.Ru != "" || aux.TranslateBuf.En != "" {
		o.Translate = make(map[string]string)
		if aux.TranslateBuf.En != "" {
			o.Translate["en"] = aux.TranslateBuf.En
		}

		if aux.TranslateBuf.Ru != "" {
			o.Translate["ru"] = aux.TranslateBuf.Ru
		}

		// Make default, or from context
		o.Display = aux.TranslateBuf.Ru
		o.Second = aux.TranslateBuf.En

		return nil
	}

	o.Display = aux.Alias.Display
	o.Second = aux.Alias.Second
	o.Translate = aux.Translate

	if _, ok := o.Translate["ru"]; !ok {
		o.Translate["ru"] = o.Display
	}

	if _, ok := o.Translate["en"]; !ok {
		o.Translate["en"] = o.Second
	}

	if o.Display == "" {
		if v, ok := o.Translate["ru"]; ok {
			o.Display = v
		}
	}

	if o.Second == "" {
		if v, ok := o.Translate["en"]; ok {
			o.Second = v
		}
	}
	return nil
}

