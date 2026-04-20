package output

import (
	"encoding/json"
	"io"

	"expr/internal/model"
)

func WriteJSON(w io.Writer, result model.ScanResult) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(result)
}
