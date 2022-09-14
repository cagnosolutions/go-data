package ui

type Attribute struct {
	Name  string
	Value string
}

type Input2 struct {
	Min string
	Max string
}

type Input struct {
	// All the global attributes for an input type can be found below.
	//
	// Specifies a unique id for an element.
	ID    string
	Type  string
	Name  string
	Value string
	// Specifies extra information about an element
	Title string
	// Specifies one or more classnames for an element
	Class string
	Style string
	// Used to store custom data private to the page or application, ie: data-*
	Data map[string]string

	// Specifies a shortcut key to activate/focus an element
	AccessKey string

	// Specifies the tabbing order of an element.
	TabIndex        int
	ContentEditable bool

	// Specifies whether the content of an element should be translated or not.
	Translate  bool
	SpellCheck bool
	ReadOnly   bool
	Disabled   bool
	Required   bool
	Hidden     bool // Specifies that an element is not yet, or is no longer, relevant.
	Draggable  bool // Specifies whether an element is draggable or not.
}

// The "size" attribute specifies the visible width, the default
// value for size is 20. ie: size="20"
//
// The input "min" and "max" attributes specify the minimum and
// maximum values for an input field. ie: min="1" max="5" or
// min="2000-01-02" max="1979-12-31"
//
// The input "multiple" attribute specifies that the user is allowed
// to enter more than one value in an input field. It is a boolean
// field. ie: multiple="true", but most of the time it is simply
// multiple.
//
// The input "pattern" attribute specifies a regular expression that
// the input field's value is checked against, when the form is
// submitted. Use the global title attribute to describe the pattern
// to help users. ie: pattern="[A-Za-z]{3}" title="3 letter co code"
//
// The input "placeholder" attribute specifies a short hint that
// describes the expected value of an input field (a sample value or a
// short description of the expected format) ie. placeholder="Full Name"
//
// The input "required" attribute specifies that an input field must be
// filled out before submitting the form. It is a boolean attribute.
//
// The input "step" attribute specifies the legal number intervals for an
// input field. Example: if step="3", legal numbers could be -3, 0, 3, 6,
// etc. Tip: This attribute can be used together with the max and min
// attributes to create a range of legal values.
//
// *The input "autofocus" attribute specifies that an input field should
// automatically get focus when the page loads. It is a boolean attribute.
//
// The input "height" and "width" attributes specify the height and width
// of an <input type="image"> element.
//
// The input "autocomplete" attribute specifies whether a form or an input
// field should have autocomplete on or off. Autocomplete allows the browser
// to predict the value. When a user starts to type in a field, the browser
// should display options to fill in the field, based on earlier typed
// values. ie: autocomplete="on", or autocomplete="off". The autocomplete
// attribute works with the <form> and certain <input> types.
//
// The input "accept" specifies a filter for what file types the user can
// pick from the file input dialog box. ie: accept="pdf", or accept="image/*"
// also mime-types are accepted as valid values.
//
// The input

type (
	Button Input

	Checkbox struct{ Input } // required
	Radio    struct{ Input } // required

	Color struct{ Input } // autocomplete

	Email struct{ Input } // size, multiple, pattern, placeholder, required, autocomplete
	File  struct{ Input } // multiple, required, accept

	Image struct{ Input } // height, width, alt

	Number   struct{ Input } // min, max, required, step
	Password struct{ Input } // size, pattern, placeholder, required, autocomplete

	Range  struct{ Input } // min, max, step, autocomplete
	Reset  struct{ Input }
	Search struct{ Input } // size, pattern, placeholder, required, autocomplete

	Submit struct{ Input }

	Tel  struct{ Input } // size, pattern, placeholder, required, autocomplete
	Text struct{ Input } // size, pattern, placeholder, autocomplete

	Url struct{ Input } // size, pattern, placeholder, required, autocomplete, src

	Date     struct{ Input } // min, max, pattern, required, step, autocomplete
	Datetime struct{ Input } // min, max, required, step, autocomplete

	Time  struct{ Input } // min, max, step
	Month struct{ Input } // min, max, step
	Week  struct{ Input } // min, max, step

	Hidden struct{ Input }
)

type Form struct {
	Text
	_ Button
}
