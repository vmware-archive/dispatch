package injectors

type injectorError struct {
	Err error `json:"err"`
}

func (err *injectorError) Error() string {
	return err.Err.Error()
}

func (err *injectorError) AsUserErrorObject() interface{} {
	return err
}
