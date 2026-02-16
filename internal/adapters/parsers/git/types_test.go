package git

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

const (
	branchMain       = "main"
	upstreamOrigin   = "origin/main"
	schemaVersion    = "https://json-schema.org/draft/2020-12/schema"
	schemaTypeObject = "object"
)

func TestStatusType(t *testing.T) {
	t.Run("can be instantiated with all fields", func(t *testing.T) {
		upstream := upstreamOrigin
		status := Status{
			Branch:    branchMain,
			Upstream:  &upstream,
			Ahead:     2,
			Behind:    1,
			Staged:    []StagedFile{{File: "file.go", Status: "added"}},
			Modified:  []string{"modified.go"},
			Deleted:   []string{"deleted.go"},
			Untracked: []string{"new.go"},
			Conflicts: []string{"conflict.go"},
			Clean:     false,
		}

		if status.Branch != branchMain {
			t.Errorf("Branch = %v, want %v", status.Branch, branchMain)
		}
		if *status.Upstream != upstreamOrigin {
			t.Errorf("Upstream = %v, want %v", *status.Upstream, upstreamOrigin)
		}
		if status.Ahead != 2 {
			t.Errorf("Ahead = %v, want %v", status.Ahead, 2)
		}
		if status.Behind != 1 {
			t.Errorf("Behind = %v, want %v", status.Behind, 1)
		}
		if len(status.Staged) != 1 || status.Staged[0].File != "file.go" {
			t.Errorf("Staged = %v, want single file.go", status.Staged)
		}
		if status.Clean {
			t.Errorf("Clean = %v, want false", status.Clean)
		}
		// Verify all slice fields are accessible
		if len(status.Modified) != 1 {
			t.Errorf("Modified length = %v, want 1", len(status.Modified))
		}
		if len(status.Deleted) != 1 {
			t.Errorf("Deleted length = %v, want 1", len(status.Deleted))
		}
		if len(status.Untracked) != 1 {
			t.Errorf("Untracked length = %v, want 1", len(status.Untracked))
		}
		if len(status.Conflicts) != 1 {
			t.Errorf("Conflicts length = %v, want 1", len(status.Conflicts))
		}
	})

	t.Run("clean status with nil upstream", func(t *testing.T) {
		status := Status{
			Branch:   branchMain,
			Upstream: nil,
			Clean:    true,
		}

		if status.Branch != branchMain {
			t.Errorf("Branch = %v, want %v", status.Branch, branchMain)
		}
		if status.Upstream != nil {
			t.Errorf("Upstream = %v, want nil", status.Upstream)
		}
		if !status.Clean {
			t.Error("Clean should be true")
		}
	})
}

func TestStagedFileType(t *testing.T) {
	tests := []struct {
		name   string
		file   string
		status string
	}{
		{"added file", "new.go", "added"},
		{"modified file", "changed.go", "modified"},
		{"deleted file", "removed.go", "deleted"},
		{"renamed file", "renamed.go", "renamed"},
		{"copied file", "copied.go", "copied"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sf := StagedFile{
				File:   tt.file,
				Status: tt.status,
			}

			if sf.File != tt.file {
				t.Errorf("File = %v, want %v", sf.File, tt.file)
			}
			if sf.Status != tt.status {
				t.Errorf("Status = %v, want %v", sf.Status, tt.status)
			}
		})
	}
}

func TestStatusJSONMarshal(t *testing.T) {
	upstream := upstreamOrigin
	status := Status{
		Branch:   "feature",
		Upstream: &upstream,
		Ahead:    1,
		Behind:   0,
		Staged: []StagedFile{
			{File: "added.go", Status: "added"},
		},
		Modified:  []string{"changed.go"},
		Deleted:   []string{},
		Untracked: []string{"new.txt"},
		Conflicts: []string{},
		Clean:     false,
	}

	data, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var unmarshaled map[string]any
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Verify JSON structure matches expected schema property names
	expectedKeys := []string{"branch", "upstream", "ahead", "behind", "staged", "modified", "deleted", "untracked", "conflicts", "clean"}
	for _, key := range expectedKeys {
		if _, exists := unmarshaled[key]; !exists {
			t.Errorf("Missing expected key in JSON: %s", key)
		}
	}

	// Verify values
	if unmarshaled["branch"] != "feature" {
		t.Errorf("branch = %v, want feature", unmarshaled["branch"])
	}
	if unmarshaled["clean"] != false {
		t.Errorf("clean = %v, want false", unmarshaled["clean"])
	}
}

func TestStatusJSONUnmarshal(t *testing.T) {
	jsonData := `{
		"branch": "main",
		"upstream": "origin/main",
		"ahead": 3,
		"behind": 2,
		"staged": [{"file": "test.go", "status": "modified"}],
		"modified": ["other.go"],
		"deleted": [],
		"untracked": [],
		"conflicts": [],
		"clean": false
	}`

	var status Status
	if err := json.Unmarshal([]byte(jsonData), &status); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if status.Branch != branchMain {
		t.Errorf("Branch = %v, want main", status.Branch)
	}
	if status.Upstream == nil || *status.Upstream != upstreamOrigin {
		t.Errorf("Upstream = %v, want origin/main", status.Upstream)
	}
	if status.Ahead != 3 {
		t.Errorf("Ahead = %v, want 3", status.Ahead)
	}
	if status.Behind != 2 {
		t.Errorf("Behind = %v, want 2", status.Behind)
	}
	if len(status.Staged) != 1 {
		t.Errorf("Staged length = %v, want 1", len(status.Staged))
	}
}

func TestStatusSchemaExists(t *testing.T) {
	// Schema path is relative to the repository root
	schemaPath := "../../../../schemas/git-status.json"
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("Failed to read schema file: %v", err)
	}

	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("Failed to parse schema JSON: %v", err)
	}

	// Verify required schema fields
	if schema["$schema"] != schemaVersion {
		t.Errorf("$schema = %v, want JSON Schema draft 2020-12", schema["$schema"])
	}
	if schema["title"] != "Git Status Output" {
		t.Errorf("title = %v, want 'Git Status Output'", schema["title"])
	}
	if schema["type"] != schemaTypeObject {
		t.Errorf("type = %v, want 'object'", schema["type"])
	}

	// Verify properties exist
	props, propsOK := schema["properties"].(map[string]any)
	if !propsOK {
		t.Fatal("properties should be an object")
	}

	expectedProps := []string{"branch", "upstream", "ahead", "behind", "staged", "modified", "deleted", "untracked", "conflicts", "clean"}
	for _, prop := range expectedProps {
		if _, exists := props[prop]; !exists {
			t.Errorf("Missing property in schema: %s", prop)
		}
	}

	// Verify required fields
	required, reqOK := schema["required"].([]any)
	if !reqOK {
		t.Fatal("required should be an array")
	}

	expectedRequired := []string{"branch", "staged", "modified", "deleted", "untracked", "conflicts", "clean"}
	requiredSet := make(map[string]bool)
	for _, r := range required {
		if s, strOK := r.(string); strOK {
			requiredSet[s] = true
		}
	}
	for _, req := range expectedRequired {
		if !requiredSet[req] {
			t.Errorf("Missing required field in schema: %s", req)
		}
	}
}

func TestStatusTypeMatchesSchema(t *testing.T) {
	// Verify the Go struct has all the expected fields with correct types
	status := Status{}
	v := reflect.TypeOf(status)

	expectedFields := map[string]string{
		"Branch":    "string",
		"Upstream":  "*string",
		"Ahead":     "int",
		"Behind":    "int",
		"Staged":    "[]git.StagedFile",
		"Modified":  "[]string",
		"Deleted":   "[]string",
		"Untracked": "[]string",
		"Conflicts": "[]string",
		"Clean":     "bool",
	}

	for fieldName, expectedType := range expectedFields {
		field, fieldOK := v.FieldByName(fieldName)
		if !fieldOK {
			t.Errorf("Missing field: %s", fieldName)
			continue
		}

		actualType := field.Type.String()
		if actualType != expectedType {
			t.Errorf("Field %s has type %s, want %s", fieldName, actualType, expectedType)
		}
	}
}

// Log Tests

func TestLogType(t *testing.T) {
	log := Log{
		Commits: []Commit{
			{
				Hash:       "abc123def456789",
				AbbrevHash: "abc123d",
				Author:     "John Doe",
				Email:      "john@example.com",
				Date:       "2024-01-15T10:30:00Z",
				Message:    "Fix bug in parser",
				Subject:    "Fix bug in parser",
			},
		},
	}

	if len(log.Commits) != 1 {
		t.Errorf("Commits length = %v, want 1", len(log.Commits))
	}
	if log.Commits[0].Hash != "abc123def456789" {
		t.Errorf("Hash = %v, want abc123def456789", log.Commits[0].Hash)
	}
	if log.Commits[0].AbbrevHash != "abc123d" {
		t.Errorf("AbbrevHash = %v, want abc123d", log.Commits[0].AbbrevHash)
	}
	if log.Commits[0].Author != "John Doe" {
		t.Errorf("Author = %v, want John Doe", log.Commits[0].Author)
	}
	if log.Commits[0].Email != "john@example.com" {
		t.Errorf("Email = %v, want john@example.com", log.Commits[0].Email)
	}
	if log.Commits[0].Date != "2024-01-15T10:30:00Z" {
		t.Errorf("Date = %v, want 2024-01-15T10:30:00Z", log.Commits[0].Date)
	}
	if log.Commits[0].Message != "Fix bug in parser" {
		t.Errorf("Message = %v, want Fix bug in parser", log.Commits[0].Message)
	}
	if log.Commits[0].Subject != "Fix bug in parser" {
		t.Errorf("Subject = %v, want Fix bug in parser", log.Commits[0].Subject)
	}
}

func TestCommitWithFiles(t *testing.T) {
	commit := Commit{
		Hash:       "abc123",
		AbbrevHash: "abc",
		Author:     "Jane",
		Email:      "jane@example.com",
		Date:       "2024-01-15",
		Message:    "Add feature",
		Subject:    "Add feature",
		Body:       "This adds a new feature\n\nWith details",
		Files: []FileChange{
			{Path: "src/main.go", Additions: 50, Deletions: 10},
		},
		Insertions: 50,
		Deletions:  10,
	}

	// Verify all required commit fields
	if commit.Hash != "abc123" {
		t.Errorf("Hash = %v, want abc123", commit.Hash)
	}
	if commit.AbbrevHash != "abc" {
		t.Errorf("AbbrevHash = %v, want abc", commit.AbbrevHash)
	}
	if commit.Author != "Jane" {
		t.Errorf("Author = %v, want Jane", commit.Author)
	}
	if commit.Email != "jane@example.com" {
		t.Errorf("Email = %v, want jane@example.com", commit.Email)
	}
	if commit.Date != "2024-01-15" {
		t.Errorf("Date = %v, want 2024-01-15", commit.Date)
	}
	if commit.Message != "Add feature" {
		t.Errorf("Message = %v, want Add feature", commit.Message)
	}
	if commit.Subject != "Add feature" {
		t.Errorf("Subject = %v, want Add feature", commit.Subject)
	}
	if commit.Body != "This adds a new feature\n\nWith details" {
		t.Errorf("Body = %v, want body text", commit.Body)
	}
	// Verify file changes
	if len(commit.Files) != 1 {
		t.Errorf("Files length = %v, want 1", len(commit.Files))
	}
	if commit.Files[0].Additions != 50 {
		t.Errorf("Additions = %v, want 50", commit.Files[0].Additions)
	}
	if commit.Files[0].Deletions != 10 {
		t.Errorf("Deletions = %v, want 10", commit.Files[0].Deletions)
	}
	if commit.Files[0].Path != "src/main.go" {
		t.Errorf("Path = %v, want src/main.go", commit.Files[0].Path)
	}
	if commit.Insertions != 50 {
		t.Errorf("Insertions = %v, want 50", commit.Insertions)
	}
	if commit.Deletions != 10 {
		t.Errorf("Deletions = %v, want 10", commit.Deletions)
	}
}

func TestLogJSONMarshal(t *testing.T) {
	log := Log{
		Commits: []Commit{
			{
				Hash:       "abc123def456",
				AbbrevHash: "abc123d",
				Author:     "Test Author",
				Email:      "test@example.com",
				Date:       "2024-01-15T10:30:00Z",
				Message:    "Test commit",
				Subject:    "Test commit",
				Files: []FileChange{
					{Path: "file.go", Additions: 10, Deletions: 5},
				},
				Insertions: 10,
				Deletions:  5,
			},
		},
	}

	data, err := json.Marshal(log)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var unmarshaled map[string]any
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Verify commits array exists
	commits, commitsOK := unmarshaled["commits"].([]any)
	if !commitsOK {
		t.Fatal("commits should be an array")
	}
	if len(commits) != 1 {
		t.Errorf("commits length = %v, want 1", len(commits))
	}

	// Verify commit structure
	commit, commitOK := commits[0].(map[string]any)
	if !commitOK {
		t.Fatal("commit should be an object")
	}

	expectedKeys := []string{"hash", "abbrevHash", "author", "email", "date", "message", "subject", "files", "insertions", "deletions"}
	for _, key := range expectedKeys {
		if _, exists := commit[key]; !exists {
			t.Errorf("Missing expected key in commit JSON: %s", key)
		}
	}
}

func TestLogSchemaExists(t *testing.T) {
	// Schema path is relative to the repository root
	schemaPath := "../../../../schemas/git-log.json"
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("Failed to read schema file: %v", err)
	}

	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("Failed to parse schema JSON: %v", err)
	}

	// Verify required schema fields
	if schema["$schema"] != schemaVersion {
		t.Errorf("$schema = %v, want JSON Schema draft 2020-12", schema["$schema"])
	}
	if schema["title"] != "Git Log Output" {
		t.Errorf("title = %v, want 'Git Log Output'", schema["title"])
	}
	if schema["type"] != schemaTypeObject {
		t.Errorf("type = %v, want 'object'", schema["type"])
	}

	// Verify properties exist
	props, propsOK := schema["properties"].(map[string]any)
	if !propsOK {
		t.Fatal("properties should be an object")
	}

	if _, exists := props["commits"]; !exists {
		t.Error("Missing property in schema: commits")
	}

	// Verify required fields
	required, reqOK := schema["required"].([]any)
	if !reqOK {
		t.Fatal("required should be an array")
	}

	found := false
	for _, r := range required {
		if s, strOK := r.(string); strOK && s == "commits" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Missing required field in schema: commits")
	}
}

func TestLogTypeMatchesSchema(t *testing.T) {
	// Verify the Go struct has all the expected fields with correct types
	log := Log{}
	v := reflect.TypeOf(log)

	field, fieldOK := v.FieldByName("Commits")
	if !fieldOK {
		t.Error("Missing field: Commits")
		return
	}

	actualType := field.Type.String()
	if actualType != "[]git.Commit" {
		t.Errorf("Field Commits has type %s, want []git.Commit", actualType)
	}
}

func TestCommitTypeMatchesSchema(t *testing.T) {
	// Verify the Commit struct has all the expected fields with correct types
	commit := Commit{}
	v := reflect.TypeOf(commit)

	expectedFields := map[string]string{
		"Hash":       "string",
		"AbbrevHash": "string",
		"Author":     "string",
		"Email":      "string",
		"Date":       "string",
		"Message":    "string",
		"Subject":    "string",
		"Body":       "string",
		"Files":      "[]git.FileChange",
		"Insertions": "int",
		"Deletions":  "int",
	}

	for fieldName, expectedType := range expectedFields {
		field, fieldOK := v.FieldByName(fieldName)
		if !fieldOK {
			t.Errorf("Missing field: %s", fieldName)
			continue
		}

		actualType := field.Type.String()
		if actualType != expectedType {
			t.Errorf("Field %s has type %s, want %s", fieldName, actualType, expectedType)
		}
	}
}

func TestFileChangeTypeMatchesSchema(t *testing.T) {
	// Verify the FileChange struct has all the expected fields with correct types
	fc := FileChange{}
	v := reflect.TypeOf(fc)

	expectedFields := map[string]string{
		"Path":      "string",
		"Additions": "int",
		"Deletions": "int",
	}

	for fieldName, expectedType := range expectedFields {
		field, fieldOK := v.FieldByName(fieldName)
		if !fieldOK {
			t.Errorf("Missing field: %s", fieldName)
			continue
		}

		actualType := field.Type.String()
		if actualType != expectedType {
			t.Errorf("Field %s has type %s, want %s", fieldName, actualType, expectedType)
		}
	}
}

// Diff Tests

func TestDiffType(t *testing.T) {
	diff := Diff{
		Files: []DiffFile{
			{
				Path:      "src/main.go",
				Status:    "modified",
				Binary:    false,
				Additions: 10,
				Deletions: 5,
			},
		},
	}

	if len(diff.Files) != 1 {
		t.Errorf("Files length = %v, want 1", len(diff.Files))
	}
	if diff.Files[0].Status != "modified" {
		t.Errorf("Status = %v, want modified", diff.Files[0].Status)
	}
}

func TestDiffFileWithRename(t *testing.T) {
	file := DiffFile{
		Path:      "new/path.go",
		OldPath:   "old/path.go",
		Status:    "renamed",
		Binary:    false,
		Additions: 0,
		Deletions: 0,
	}

	if file.Path != "new/path.go" {
		t.Errorf("Path = %v, want new/path.go", file.Path)
	}
	if file.OldPath != "old/path.go" {
		t.Errorf("OldPath = %v, want old/path.go", file.OldPath)
	}
	if file.Status != "renamed" {
		t.Errorf("Status = %v, want renamed", file.Status)
	}
	if file.Binary {
		t.Errorf("Binary = %v, want false", file.Binary)
	}
	if file.Additions != 0 {
		t.Errorf("Additions = %v, want 0", file.Additions)
	}
	if file.Deletions != 0 {
		t.Errorf("Deletions = %v, want 0", file.Deletions)
	}
}

func TestDiffHunkType(t *testing.T) {
	hunk := DiffHunk{
		OldStart: 10,
		OldLines: 5,
		NewStart: 10,
		NewLines: 7,
		Header:   "@@ -10,5 +10,7 @@ func main()",
		Lines: []DiffLine{
			{Type: "context", Content: " func main() {"},
			{Type: "delete", Content: "-    old line"},
			{Type: "add", Content: "+    new line"},
		},
	}

	if len(hunk.Lines) != 3 {
		t.Errorf("Lines length = %v, want 3", len(hunk.Lines))
	}
	if hunk.OldStart != 10 {
		t.Errorf("OldStart = %v, want 10", hunk.OldStart)
	}
	if hunk.OldLines != 5 {
		t.Errorf("OldLines = %v, want 5", hunk.OldLines)
	}
	if hunk.NewStart != 10 {
		t.Errorf("NewStart = %v, want 10", hunk.NewStart)
	}
	if hunk.NewLines != 7 {
		t.Errorf("NewLines = %v, want 7", hunk.NewLines)
	}
	if hunk.Header != "@@ -10,5 +10,7 @@ func main()" {
		t.Errorf("Header = %v, want @@ -10,5 +10,7 @@ func main()", hunk.Header)
	}
}

func TestDiffLineType(t *testing.T) {
	tests := []struct {
		name      string
		lineType  string
		content   string
		oldLineNo *int
		newLineNo *int
	}{
		{"context line", "context", " unchanged", intPtr(10), intPtr(10)},
		{"add line", "add", "+new line", nil, intPtr(11)},
		{"delete line", "delete", "-old line", intPtr(11), nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			line := DiffLine{
				Type:      tt.lineType,
				Content:   tt.content,
				OldLineNo: tt.oldLineNo,
				NewLineNo: tt.newLineNo,
			}

			if line.Type != tt.lineType {
				t.Errorf("Type = %v, want %v", line.Type, tt.lineType)
			}
			if line.Content != tt.content {
				t.Errorf("Content = %v, want %v", line.Content, tt.content)
			}
			// Verify line numbers match expected values
			if tt.oldLineNo != nil {
				if line.OldLineNo == nil || *line.OldLineNo != *tt.oldLineNo {
					t.Errorf("OldLineNo = %v, want %v", line.OldLineNo, *tt.oldLineNo)
				}
			} else if line.OldLineNo != nil {
				t.Errorf("OldLineNo = %v, want nil", *line.OldLineNo)
			}
			if tt.newLineNo != nil {
				if line.NewLineNo == nil || *line.NewLineNo != *tt.newLineNo {
					t.Errorf("NewLineNo = %v, want %v", line.NewLineNo, *tt.newLineNo)
				}
			} else if line.NewLineNo != nil {
				t.Errorf("NewLineNo = %v, want nil", *line.NewLineNo)
			}
		})
	}
}

func intPtr(i int) *int {
	return &i
}

func TestDiffJSONMarshal(t *testing.T) {
	diff := Diff{
		Files: []DiffFile{
			{
				Path:      "src/main.go",
				Status:    "modified",
				Binary:    false,
				Additions: 10,
				Deletions: 5,
				Hunks: []DiffHunk{
					{
						OldStart: 1,
						OldLines: 5,
						NewStart: 1,
						NewLines: 7,
						Header:   "@@ -1,5 +1,7 @@",
						Lines: []DiffLine{
							{Type: "context", Content: " line1"},
							{Type: "add", Content: "+new line"},
						},
					},
				},
			},
		},
	}

	data, err := json.Marshal(diff)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var unmarshaled map[string]any
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	// Verify files array exists
	files, filesOK := unmarshaled["files"].([]any)
	if !filesOK {
		t.Fatal("files should be an array")
	}
	if len(files) != 1 {
		t.Errorf("files length = %v, want 1", len(files))
	}

	// Verify file structure
	file, fileOK := files[0].(map[string]any)
	if !fileOK {
		t.Fatal("file should be an object")
	}

	expectedKeys := []string{"path", "status", "binary", "additions", "deletions", "hunks"}
	for _, key := range expectedKeys {
		if _, exists := file[key]; !exists {
			t.Errorf("Missing expected key in file JSON: %s", key)
		}
	}
}

func TestDiffSchemaExists(t *testing.T) {
	// Schema path is relative to the repository root
	schemaPath := "../../../../schemas/git-diff.json"
	data, err := os.ReadFile(schemaPath)
	if err != nil {
		t.Fatalf("Failed to read schema file: %v", err)
	}

	var schema map[string]any
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("Failed to parse schema JSON: %v", err)
	}

	// Verify required schema fields
	if schema["$schema"] != schemaVersion {
		t.Errorf("$schema = %v, want JSON Schema draft 2020-12", schema["$schema"])
	}
	if schema["title"] != "Git Diff Output" {
		t.Errorf("title = %v, want 'Git Diff Output'", schema["title"])
	}
	if schema["type"] != schemaTypeObject {
		t.Errorf("type = %v, want 'object'", schema["type"])
	}

	// Verify properties exist
	props, propsOK := schema["properties"].(map[string]any)
	if !propsOK {
		t.Fatal("properties should be an object")
	}

	if _, exists := props["files"]; !exists {
		t.Error("Missing property in schema: files")
	}

	// Verify required fields
	required, reqOK := schema["required"].([]any)
	if !reqOK {
		t.Fatal("required should be an array")
	}

	found := false
	for _, r := range required {
		if s, strOK := r.(string); strOK && s == "files" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Missing required field in schema: files")
	}
}

func TestDiffTypeMatchesSchema(t *testing.T) {
	// Verify the Go struct has all the expected fields with correct types
	diff := Diff{}
	v := reflect.TypeOf(diff)

	field, fieldOK := v.FieldByName("Files")
	if !fieldOK {
		t.Error("Missing field: Files")
		return
	}

	actualType := field.Type.String()
	if actualType != "[]git.DiffFile" {
		t.Errorf("Field Files has type %s, want []git.DiffFile", actualType)
	}
}

func TestDiffFileTypeMatchesSchema(t *testing.T) {
	// Verify the DiffFile struct has all the expected fields with correct types
	df := DiffFile{}
	v := reflect.TypeOf(df)

	expectedFields := map[string]string{
		"Path":      "string",
		"OldPath":   "string",
		"Status":    "string",
		"Binary":    "bool",
		"Additions": "int",
		"Deletions": "int",
		"Hunks":     "[]git.DiffHunk",
	}

	for fieldName, expectedType := range expectedFields {
		field, fieldOK := v.FieldByName(fieldName)
		if !fieldOK {
			t.Errorf("Missing field: %s", fieldName)
			continue
		}

		actualType := field.Type.String()
		if actualType != expectedType {
			t.Errorf("Field %s has type %s, want %s", fieldName, actualType, expectedType)
		}
	}
}

func TestDiffHunkTypeMatchesSchema(t *testing.T) {
	// Verify the DiffHunk struct has all the expected fields with correct types
	dh := DiffHunk{}
	v := reflect.TypeOf(dh)

	expectedFields := map[string]string{
		"OldStart": "int",
		"OldLines": "int",
		"NewStart": "int",
		"NewLines": "int",
		"Header":   "string",
		"Lines":    "[]git.DiffLine",
	}

	for fieldName, expectedType := range expectedFields {
		field, fieldOK := v.FieldByName(fieldName)
		if !fieldOK {
			t.Errorf("Missing field: %s", fieldName)
			continue
		}

		actualType := field.Type.String()
		if actualType != expectedType {
			t.Errorf("Field %s has type %s, want %s", fieldName, actualType, expectedType)
		}
	}
}

func TestDiffLineTypeMatchesSchema(t *testing.T) {
	// Verify the DiffLine struct has all the expected fields with correct types
	dl := DiffLine{}
	v := reflect.TypeOf(dl)

	expectedFields := map[string]string{
		"Type":      "string",
		"Content":   "string",
		"OldLineNo": "*int",
		"NewLineNo": "*int",
	}

	for fieldName, expectedType := range expectedFields {
		field, fieldOK := v.FieldByName(fieldName)
		if !fieldOK {
			t.Errorf("Missing field: %s", fieldName)
			continue
		}

		actualType := field.Type.String()
		if actualType != expectedType {
			t.Errorf("Field %s has type %s, want %s", fieldName, actualType, expectedType)
		}
	}
}
