package queue

import (
	"encoding/json"
	"testing"
)

func TestFontJobJSONRoundtrip(t *testing.T) {
	original := FontJob{
		FontID:           "font-abc-123",
		UserID:           "user-456",
		TemplateScanPath: "/storage/scans/template.png",
		OutputPath:       "/storage/fonts/output.ttf",
	}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("failed to marshal FontJob: %v", err)
	}

	var decoded FontJob
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("failed to unmarshal FontJob: %v", err)
	}

	if decoded.FontID != original.FontID {
		t.Errorf("FontID: got %q, want %q", decoded.FontID, original.FontID)
	}
	if decoded.UserID != original.UserID {
		t.Errorf("UserID: got %q, want %q", decoded.UserID, original.UserID)
	}
	if decoded.TemplateScanPath != original.TemplateScanPath {
		t.Errorf("TemplateScanPath: got %q, want %q", decoded.TemplateScanPath, original.TemplateScanPath)
	}
	if decoded.OutputPath != original.OutputPath {
		t.Errorf("OutputPath: got %q, want %q", decoded.OutputPath, original.OutputPath)
	}
}

func TestFontJobJSONFieldNames(t *testing.T) {
	job := FontJob{
		FontID:           "f1",
		UserID:           "u1",
		TemplateScanPath: "/scan.png",
		OutputPath:       "/out.ttf",
	}

	data, err := json.Marshal(job)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var raw map[string]string
	err = json.Unmarshal(data, &raw)
	if err != nil {
		t.Fatalf("failed to unmarshal to map: %v", err)
	}

	expectedKeys := []string{"font_id", "user_id", "template_scan_path", "output_path"}
	for _, key := range expectedKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("expected JSON key %q not found in %v", key, raw)
		}
	}
}
