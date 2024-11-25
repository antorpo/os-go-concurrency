package pipeline

import (
	"fmt"
	"reflect"
)

type (
	HColor        string
	FontName      string
	FontStyle     string
	TextAlignment string
	Size          int

	Include string
	Skin    string

	Canvas struct {
		Background HColor `skinparam:"backgroundColor"`
		Shadowing  bool   `skinparam:"shadowing"`

		DefaultFont      FontName  `skinparam:"defaultFontName"`
		DefaultFontColor HColor    `skinparam:"defaultFontColor"`
		DefaultFontSize  Size      `skinparam:"defaultFontSize"`
		DefaultFontStyle FontStyle `skinparam:"defaultFontStyle"`
		HyperLink        HColor    `skinparam:"HyperlinkColor"`
	}

	StageStyles struct {
		BackgroundColor  HColor    `skinparam:"ActivityBackgroundColor"`
		FontName         FontName  `skinparam:"ActivityFontName"`
		FontColor        HColor    `skinparam:"ActivityFontColor"`
		PackageFontColor HColor    `variable:"packageColor"`
		FontSize         Size      `skinparam:"ActivityFontSize"`
		FontStyle        FontStyle `skinparam:"ActivityFontStyle"`
		BorderColor      HColor    `skinparam:"ActivityBorderColor"`
		Border           Size      `skinparam:"ActivityBorderThickness"`
	}

	SplicerStyles struct {
		FontColor HColor `variable:"splicerFontColor"`
		FontSize  Size   `variable:"splicerFontSize"`
	}

	TransactionalStyles struct {
		BackgroundColor HColor    `skinparam:"PartitionBackgroundColor"`
		FontName        FontName  `skinparam:"PartitionFontName"`
		FontColor       HColor    `skinparam:"PartitionFontColor"`
		FontSize        Size      `skinparam:"PartitionFontSize"`
		FontStyle       FontStyle `skinparam:"PartitionFontStyle"`
		BorderColor     HColor    `skinparam:"PartitionBorderColor"`
		Border          Size      `skinparam:"PartitionBorderThickness"`
	}

	ConnectorsStyles struct {
		ArrowColor            HColor        `skinparam:"ArrowColor"`
		ArrowFontColor        HColor        `skinparam:"ArrowFontColor"`
		ArrowFontName         FontName      `skinparam:"ArrowFontName"`
		ArrowFontSize         Size          `skinparam:"ArrowFontSize"`
		ArrowFontStyle        FontStyle     `skinparam:"ArrowFontStyle"`
		ArrowMessageAlignment TextAlignment `skinparam:"ArrowMessageAlignment"`

		Bar HColor `skinparam:"ActivityBarColor"`
	}

	FlowControlStyles struct {
		BackgroundColor HColor    `skinparam:"ActivityDiamondBackgroundColor"`
		FontName        FontName  `skinparam:"ActivityDiamondFontName"`
		FontColor       HColor    `skinparam:"ActivityDiamondFontColor"`
		FontSize        Size      `skinparam:"ActivityDiamondFontSize"`
		FontStyle       FontStyle `skinparam:"ActivityDiamondFontStyle"`
		BorderColor     HColor    `skinparam:"ActivityDiamondBorderColor"`
	}

	NoteStyles struct {
		BackgroundColor HColor        `skinparam:"NoteBackgroundColor"`
		BorderColor     HColor        `skinparam:"NoteBorderColor"`
		Border          Size          `skinparam:"NoteBorderThickness"`
		FontColor       HColor        `skinparam:"NoteFontColor"`
		FontName        FontName      `skinparam:"NoteFontName"`
		FontSize        Size          `skinparam:"NoteFontSize"`
		FontStyle       FontStyle     `skinparam:"NoteFontStyle"`
		Shadowing       bool          `skinparam:"NoteShadowing"`
		TextAlignment   TextAlignment `skinparam:"NoteTextAlignment"`
	}

	TraceTagStyles struct {
		FontColor HColor `variable:"traceTagColor"`
	}

	ErrorStyles struct {
		FontColor HColor `variable:"errorColor"`
		FontSize  Size   `variable:"errorSize"`
	}

	BlueprintSkinParams struct {
		Include
		Canvas        Canvas
		Stages        StageStyles
		Splicers      SplicerStyles
		Transactional TransactionalStyles
		Connectors    ConnectorsStyles
		FlowControl   FlowControlStyles
		Notes         NoteStyles
	}

	TracerSkinParams struct {
		Include
		Canvas        Canvas
		Stages        StageStyles
		Transactional TransactionalStyles
		Connectors    ConnectorsStyles
		Notes         NoteStyles
		Tags          TraceTagStyles
		Errors        ErrorStyles
	}
)

func (s BlueprintSkinParams) Build() Skin {
	return skin(s.Include, parse(s), func() bool {
		return reflect.DeepEqual(s, BlueprintSkinParams{Include: s.Include})
	})
}

func (s TracerSkinParams) Build() Skin {
	return skin(s.Include, parse(s), func() bool {
		return reflect.DeepEqual(s, TracerSkinParams{Include: s.Include})
	})
}

func skin(include Include, parsed string, zFn func() bool) Skin {
	var out string

	if include != "" {
		out += fmt.Sprintf("!include %s\n", include)
	}

	if !zFn() {
		out += parsed
	}

	return Skin(out)
}

func parse(v interface{}) string {
	elemType := reflect.TypeOf(v)
	value := reflect.ValueOf(v)

	var out string

	for i := 0; i < elemType.NumField(); i++ {
		fieldType := elemType.Field(i)
		fieldValue := value.Field(i)

		if fieldType.Type.Kind() == reflect.Struct {
			out += parse(fieldValue.Interface())
			continue
		}

		if isZero(fieldValue) {
			continue
		}

		param := elemType.Field(i).Tag.Get("skinparam")
		if len(param) > 0 {
			out += fmt.Sprintf("skinparam %v %v \n", param, fieldValue.Interface())
		}

		variable := elemType.Field(i).Tag.Get("variable")
		if len(variable) > 0 {
			out += fmt.Sprintf("!$%v = \"%v\" \n", variable, fieldValue.Interface())
		}
	}

	return out
}

func isZero(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.String:
		return v.Len() == 0
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	}

	return false
}

const (
	font32 = 32
	font24 = 24
	font16 = 16
	font13 = 13
	font12 = 12
	font11 = 11
	font10 = 10
	font9  = 9

	border3  = 3
	border2  = 2
	border1  = 1
	noBorder = 0
)

var (
	DemoBlueprintSkin = BlueprintSkinParams{
		Canvas: Canvas{
			Background:       "#ffffff",
			Shadowing:        false,
			DefaultFontColor: "#3c415e",
		},
		Stages: StageStyles{
			BackgroundColor:  "#fafafa",
			FontName:         "Tahoma",
			FontColor:        "#1cb3c8",
			FontSize:         font13,
			FontStyle:        "bold",
			BorderColor:      "#ffffff",
			Border:           noBorder,
			PackageFontColor: "#738598",
		},
		FlowControl: FlowControlStyles{
			BackgroundColor: "#1cb3c8",
			FontName:        "Tahoma",
			FontColor:       "#3c415e",
			FontSize:        font13,
			FontStyle:       "bold italic",
			BorderColor:     "#3c415e",
		},
		Notes: NoteStyles{
			BackgroundColor: "#ffffff",
			BorderColor:     "#ffffff",
			Border:          border1,
			FontColor:       "#3c415e",
			FontName:        "Arial",
			FontSize:        font11,
			FontStyle:       "bold",
			Shadowing:       true,
			TextAlignment:   "left",
		},
		Connectors: ConnectorsStyles{
			ArrowColor:            "#738598",
			ArrowFontColor:        "#3c415e",
			ArrowFontSize:         font10,
			ArrowFontStyle:        "italic",
			ArrowMessageAlignment: "center",
			Bar:                   "#3c415e",
		},
		Transactional: TransactionalStyles{
			BackgroundColor: "#ffffff",
			FontColor:       "#3c415e",
			FontName:        "Arial",
			FontSize:        font10,
			FontStyle:       "regular",
			BorderColor:     "#738598",
			Border:          border1,
		},
	}
	DemoTraceSkin = TracerSkinParams{
		Canvas: Canvas{
			Background:       "#222831",
			Shadowing:        true,
			DefaultFontColor: "#eeeeee",
			HyperLink:        "#00adb5",
		},
		Stages: StageStyles{
			BackgroundColor:  "#393e46",
			FontName:         "Tahoma",
			FontColor:        "#00adb5",
			FontSize:         font12,
			BorderColor:      "#222831",
			Border:           border1,
			PackageFontColor: "#222831",
		},
		Transactional: TransactionalStyles{
			BackgroundColor: "#222831",
			FontName:        "Tahoma",
			FontColor:       "#00adb5",
			FontSize:        font12,
			FontStyle:       "bold",
			BorderColor:     "#222831",
			Border:          border1,
		},
		Notes: NoteStyles{
			BackgroundColor: "#393e46",
			BorderColor:     "#00adb5",
			Border:          border1,
			FontColor:       "#eeeeee",
			FontName:        "Arial",
			FontSize:        font12,
			FontStyle:       "bold",
			Shadowing:       true,
			TextAlignment:   "left",
		},
		Connectors: ConnectorsStyles{
			ArrowColor:            "#393e46",
			ArrowFontColor:        "#00adb5",
			ArrowFontSize:         font10,
			ArrowFontStyle:        "italic",
			ArrowMessageAlignment: "center",
			Bar:                   "#393e46",
		},
		Tags: TraceTagStyles{
			FontColor: "#eeeeee",
		},
		Errors: ErrorStyles{
			FontColor: "#ffffff",
			FontSize:  font16,
		},
	}

	ParanoidBlueprintSkin = BlueprintSkinParams{
		Canvas: Canvas{
			Background: "#070607",
			Shadowing:  false,
		},
		Stages: StageStyles{
			BackgroundColor:  "#DC5B5F",
			FontName:         "Tahoma",
			FontColor:        "#070607",
			FontSize:         font12,
			FontStyle:        "bold",
			BorderColor:      "#541823",
			Border:           border3,
			PackageFontColor: "#196597",
		},
		FlowControl: FlowControlStyles{
			BackgroundColor: "#339FB9",
			FontName:        "Tahoma",
			FontColor:       "#070607",
			FontSize:        font13,
			FontStyle:       "bold italic",
			BorderColor:     "#196597",
		},
		Notes: NoteStyles{
			BackgroundColor: "#070607",
			BorderColor:     "#070607",
			Border:          border1,
			FontColor:       "#DC5B5F",
			FontName:        "Arial",
			FontSize:        font11,
			FontStyle:       "bold",
			Shadowing:       true,
			TextAlignment:   "left",
		},
		Connectors: ConnectorsStyles{
			ArrowColor:            "#196597",
			ArrowFontColor:        "#339FB9",
			ArrowFontSize:         font10,
			ArrowFontStyle:        "italic",
			ArrowMessageAlignment: "center",
			Bar:                   "#339FB9",
		},
	}
	RemoteParanoidBlueprintSkin = BlueprintSkinParams{
		Include: "https://raw.githubusercontent.com/pecheverria/pipeline-themes/main/paranoid-blueprint.skin",
	}

	TeaBluePrintSkin = BlueprintSkinParams{
		Canvas: Canvas{
			Background:       "#fdf6e3",
			Shadowing:        false,
			DefaultFontColor: "#33322E",
			HyperLink:        "#76736a",
			DefaultFontSize:  font12,
			DefaultFontStyle: "normal",
		},
		Stages: StageStyles{
			BackgroundColor:  "#eee8d5",
			FontName:         "Tahoma",
			FontColor:        "#33322E",
			FontSize:         font10,
			FontStyle:        "bold",
			BorderColor:      "#33322f",
			Border:           border2,
			PackageFontColor: "#96a4a3",
		},
		Splicers: SplicerStyles{
			FontColor: "#33322E",
			FontSize:  font24,
		},
		FlowControl: FlowControlStyles{
			BackgroundColor: "#33322E",
			FontName:        "Tahoma",
			FontColor:       "#fdf6e3",
			FontSize:        font13,
			FontStyle:       "bold italic",
			BorderColor:     "#33322f",
		},
		Notes: NoteStyles{
			BackgroundColor: "#fdf6e3",
			BorderColor:     "#fdf6e3",
			Border:          noBorder,
			FontColor:       "#657b83",
			FontName:        "Arial",
			FontSize:        font9,
			FontStyle:       "normal",
			Shadowing:       false,
			TextAlignment:   "left",
		},
		Connectors: ConnectorsStyles{
			ArrowColor:            "#33322f",
			ArrowFontColor:        "#33322f",
			ArrowFontSize:         font10,
			ArrowFontStyle:        "italic",
			ArrowMessageAlignment: "center",
			Bar:                   "#33322f",
		},
	}
	RemoteTeaBluePrintSkin = BlueprintSkinParams{
		Include: "https://raw.githubusercontent.com/pecheverria/pipeline-themes/main/tea-blueprint.skin",
	}

	TeaTracerSkin = TracerSkinParams{
		Canvas: Canvas{
			Background:       "#fdf6e3",
			Shadowing:        true,
			DefaultFontColor: "#33322E",
			HyperLink:        "#76736a",
			DefaultFontSize:  font12,
			DefaultFontStyle: "normal",
		},
		Stages: StageStyles{
			BackgroundColor:  "#eee8d5",
			FontName:         "Tahoma",
			FontColor:        "#33322E",
			FontSize:         font10,
			FontStyle:        "bold",
			BorderColor:      "#33322f",
			Border:           border2,
			PackageFontColor: "#96a4a3",
		},
		Transactional: TransactionalStyles{
			BackgroundColor: "#eee8d5",
			FontName:        "Tahoma",
			FontColor:       "#33322E",
			FontSize:        font10,
			FontStyle:       "bold",
			BorderColor:     "#eee8d5",
			Border:          border1,
		},
		Notes: NoteStyles{
			BackgroundColor: "#fdf6e3",
			BorderColor:     "#657b83",
			Border:          border1,
			FontColor:       "#33322E",
			FontName:        "Arial",
			FontSize:        font12,
			FontStyle:       "bold",
			Shadowing:       true,
			TextAlignment:   "left",
		},
		Connectors: ConnectorsStyles{
			ArrowColor:            "#33322f",
			ArrowFontColor:        "#33322f",
			ArrowFontSize:         font10,
			ArrowFontStyle:        "italic",
			ArrowMessageAlignment: "center",
			Bar:                   "#33322f",
		},
		Tags: TraceTagStyles{
			FontColor: "#33322E",
		},
		Errors: ErrorStyles{
			FontColor: "#33322E",
			FontSize:  font16,
		},
	}
	RemoteTeaTracerSkin = BlueprintSkinParams{
		Include: "https://raw.githubusercontent.com/pecheverria/pipeline-themes/main/tea-tracer.skin",
	}

	ShadesOfBlueprintSkin = BlueprintSkinParams{
		Canvas: Canvas{
			Background:       "#dfe2e2",
			Shadowing:        false,
			DefaultFontColor: "#3c415e",
		},
		Stages: StageStyles{
			BackgroundColor:  "#3c415e",
			FontName:         "Tahoma",
			FontColor:        "#dfe2e2",
			FontSize:         font12,
			FontStyle:        "bold",
			BorderColor:      "#1cb3c8",
			Border:           border3,
			PackageFontColor: "#738598",
		},
		Splicers: SplicerStyles{
			FontColor: "#3c415e",
			FontSize:  font32,
		},
		FlowControl: FlowControlStyles{
			BackgroundColor: "#1cb3c8",
			FontName:        "Tahoma",
			FontColor:       "#3c415e",
			FontSize:        font13,
			FontStyle:       "bold italic",
			BorderColor:     "#3c415e",
		},
		Notes: NoteStyles{
			BackgroundColor: "#dfe2e2",
			BorderColor:     "#dfe2e2",
			Border:          border1,
			FontColor:       "#3c415e",
			FontName:        "Arial",
			FontSize:        font11,
			FontStyle:       "bold",
			Shadowing:       true,
			TextAlignment:   "left",
		},
		Connectors: ConnectorsStyles{
			ArrowColor:            "#738598",
			ArrowFontColor:        "#3c415e",
			ArrowFontSize:         font10,
			ArrowFontStyle:        "italic",
			ArrowMessageAlignment: "center",
			Bar:                   "#3c415e",
		},
		Transactional: TransactionalStyles{
			BackgroundColor: "#dfe2e2",
			FontName:        "Tahoma",
			FontColor:       "#3c415e",
			FontSize:        font16,
			FontStyle:       "bold",
			BorderColor:     "#738598",
			Border:          border1,
		},
	}
	UrdinaBlueprintSkin       = ShadesOfBlueprintSkin
	RemoteUrdinaBlueprintSkin = BlueprintSkinParams{
		Include: "https://raw.githubusercontent.com/pecheverria/pipeline-themes/main/urdina-blueprint.skin",
	}

	NightNeonTracer = TracerSkinParams{
		Canvas: Canvas{
			Background:       "#222831",
			Shadowing:        true,
			DefaultFontColor: "#eeeeee",
			HyperLink:        "#00adb5",
		},
		Stages: StageStyles{
			BackgroundColor:  "#393e46",
			FontName:         "Tahoma",
			FontColor:        "#00adb5",
			FontSize:         font12,
			FontStyle:        "bold",
			BorderColor:      "#222831",
			Border:           border1,
			PackageFontColor: "#222831",
		},
		Transactional: TransactionalStyles{
			BackgroundColor: "#222831",
			FontName:        "Tahoma",
			FontColor:       "#00adb5",
			FontSize:        font12,
			FontStyle:       "bold",
			BorderColor:     "#222831",
			Border:          border1,
		},
		Notes: NoteStyles{
			BackgroundColor: "#393e46",
			BorderColor:     "#00adb5",
			Border:          border1,
			FontColor:       "#eeeeee",
			FontName:        "Arial",
			FontSize:        font12,
			FontStyle:       "bold",
			Shadowing:       true,
			TextAlignment:   "left",
		},
		Connectors: ConnectorsStyles{
			ArrowColor:            "#393e46",
			ArrowFontColor:        "#00adb5",
			ArrowFontSize:         font10,
			ArrowFontStyle:        "italic",
			ArrowMessageAlignment: "center",
			Bar:                   "#393e46",
		},
		Tags: TraceTagStyles{
			FontColor: "#eeeeee",
		},
		Errors: ErrorStyles{
			FontColor: "#FAFAFA",
			FontSize:  font16,
		},
	}
	RemoteNightNeonTracer = BlueprintSkinParams{
		Include: "https://raw.githubusercontent.com/pecheverria/pipeline-themes/main/nightneon-tracer.skin",
	}
)
