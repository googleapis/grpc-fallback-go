package server

type rawCodec struct{}

func (rawCodec) Marshal(v interface{}) ([]byte, error) {
	return v.([]byte), nil
}

func (rawCodec) Unmarshal(data []byte, v interface{}) error {
	*(v.(*[]byte)) = data
	return nil
}

func (rawCodec) Name() string {
	return "raw"
}
