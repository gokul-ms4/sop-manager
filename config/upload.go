package config
import (
    "bytes"
    "fmt"
    "io"
    "net/http"
    "encoding/json"
    "mime/multipart"
    "encoding/hex"
    "os"
    "strings"
    "path/filepath"
    "time"
)
var allowedMagicBytes = map[string][]string{
    // Images
    "image/jpeg": {"FFD8FF"},
    "image/png":  {"89504E47"},
    "image/gif":  {"47494638"},
    "application/pdf": {"25504446"},
    // Vector / CAD
    "image/svg+xml": {"3C3F786D6C", "3C737667"}, // <?xml / <svg
    "application/acad": {"41433130"},            // DWG "AC10"
    "application/vnd.ms-project": {},
    "application/octet-stream": {},            // MPP (no fixed signature, rely on MIME)
    // Video
    "video/mp4": {"00000018 66747970", "00000020 66747970"}, // ftyp
    "video/x-msvideo": {"52494646"}, // RIFF
    "video/x-matroska": {"1A45DFA3"}, // MKV/WebM
    "video/webm":      {"1A45DFA3"},
    // Audio
    "audio/mpeg": {"494433", "FFF1", "FFF9"}, // MP3
    "audio/wav": {"52494646"},                // RIFF → check "WAVE" chunk
    "audio/ogg": {"4F676753"},                // OggS
    "audio/flac": {"664C6143"},               // fLaC
    "audio/aac": {"FFF1", "FFF9"},            // AAC
    "audio/webm": {"1A45DFA3"},               // WebM audio-only
    // Excel
    "application/vnd.ms-excel": {"D0CF11E0A1B11AE1"}, // XLS (binary)
    "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet": {"504B0304"}, // XLSX (ZIP)
}
func isValidFileType(file multipart.File) (string, bool) {
    buf := make([]byte, 512)
    n, _ := file.Read(buf)
    file.Seek(0, io.SeekStart)
    mimeType := http.DetectContentType(buf[:n])
    allowedSigs, ok := allowedMagicBytes[mimeType]
    if !ok {
        return mimeType, false
    }
    if len(allowedSigs) > 0 {
        hexBytes := strings.ToUpper(hex.EncodeToString(buf[:n]))
        for _, sig := range allowedSigs {
            sig = strings.ReplaceAll(sig, " ", "") // clean spaces
            if strings.HasPrefix(hexBytes, sig) {
                return mimeType, true
            }
        }
        return mimeType, false
    }
    return mimeType, true
}
func UploadFile(file *multipart.FileHeader) (string, error) {
    src, err := file.Open()
    if err != nil {
        return "", fmt.Errorf("unable to open file")
    }
    defer src.Close()

    mimeType, valid := isValidFileType(src)
    if !valid {
        return "", fmt.Errorf("file type not allowed: %s", mimeType)
    }

    // Seek back to start after MIME type check consumed the reader
    if _, err := src.Seek(0, io.SeekStart); err != nil {
        return "", fmt.Errorf("unable to reset file reader")
    }

    // Ensure uploads directory exists
    if err := os.MkdirAll("./uploads", 0755); err != nil {
        return "", fmt.Errorf("unable to create uploads directory")
    }

    filename := fmt.Sprintf("project_%d%s", time.Now().UnixNano(), filepath.Ext(file.Filename))
    dstPath := filepath.Join("./uploads", filename)

    dst, err := os.Create(dstPath)
    if err != nil {
        return "", fmt.Errorf("unable to create destination file: %w", err)
    }
    defer dst.Close()

    if _, err := io.Copy(dst, src); err != nil {
        return "", fmt.Errorf("unable to write file: %w", err)
    }

    return filename, nil
}
func DeleteFile(filename string) error {
    fullPath := "./uploads/" + filename // adjust if needed
    fmt.Println("Deleting file:", fullPath) // debug
    if _, err := os.Stat(fullPath); os.IsNotExist(err) {
        return fmt.Errorf("file does not exist: %s", fullPath)
    }
    if err := os.Remove(fullPath); err != nil {
        return fmt.Errorf("failed to delete file: %v", err)
    }
    return nil
}
type SMSUsers struct {
    Username    string `json:"username"`
    PhoneNumber string `json:"phone_number"`
    Email       string `json:"email"`
    Password    string `json:"password"`
}
func SendSMS(user SMSUsers, otp string, smsType string) error {
    url := "https://control.msg91.com/api/v5/flow"
    if user.PhoneNumber == "" {
        return fmt.Errorf("user does not have a phone number")
    }
    var payload map[string]interface{}
    switch smsType {
    case "welcome":
        payload = map[string]interface{}{
            "flow_id"     :  "687e02cfd6fc054f4b180222",
            "sender"  :  "CLODHS",
            "mobiles" :  user.PhoneNumber,
            "name"        :  user.Username,
            "var"     :  "KITFRA",
            "username"    :  user.Email,
            "password"    :  user.Password,
        }
    case "login_otp":
        payload = map[string]interface{}{
            "flow_id"     :  "68777d4fd6fc0532a9310ef2",
            "sender"  :  "CLODHS",
            "mobiles" :  user.PhoneNumber,
            "name"      :  user.Username,
            "otp"       :  otp,
            "product"   : "KITFRA",
        }
    case "otp":
        payload = map[string]interface{}{
            "flow_id"     :  "68777d4fd6fc0532a9310ef2",
            "sender"  :  "CLODHS",
            "mobiles" :  user.PhoneNumber,
            "name"      :  user.Username,
            "otp"       :  otp,
            "product"   : "KITFRA",
        }
    case "reset password":
        payload = map[string]interface{}{
            "flow_id"     :  "687e0268d6fc052c6b7cca62",
            "sender"  :  "CLODHS",
            "mobiles" :  user.PhoneNumber,
            "name"        :  user.Username,
            "username"    :  user.Email,
            "password"    :  user.Password,
        }
    default:
        return fmt.Errorf("unsupported SMS type: %s", smsType)
    }
    jsonPayload, err := json.Marshal(payload)
    if err != nil {
        return err
    }
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
    if err != nil {
        return err
    }
    req.Header.Set("authkey", "457286AXLfk8i9687e0f42P1")
    req.Header.Set("Content-Type", "application/json")
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    fmt.Println("Response Status:", resp.Status)
    return nil
}
type SMSTicketUsers struct {
    Username    string `json:"username"`
    PhoneNumber string `json:"phone_number"`
    TicketID    string `json:"ticket_id"`
    Project     string `json:"project"`
    Task        string `json:"task"`
    Subtask     string `json:"subtask"`
    Link        string `json:"link"`
}
func SendTicketSMS(user SMSTicketUsers) error {
    url := "https://control.msg91.com/api/v5/flow"
    frontendURL := os.Getenv("FRONTEND_URL")
    supportLink :=frontendURL+"/"+user.Link
    //supportLink := user.Link
    if user.PhoneNumber == "" {
        return fmt.Errorf("user does not have a phone number")
    }
    payload := map[string]interface{}{
        "flow_id"     :  "687e035cd6fc05092450c042",
        "sender"  :  "CLODHS",
        "mobiles" :  user.PhoneNumber,
        "name"      :  user.Username,
        "project"   :  user.Project,
        "ticket"    :  user.TicketID,
        "task"      :  user.Task, 
        "subtask"   :  user.Subtask,
        "details"   :  supportLink,
    }
    jsonPayload, err := json.Marshal(payload)
    if err != nil {
        return err
    }
    req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonPayload))
    if err != nil {
        return err
    }
    req.Header.Set("authkey", "457286AXLfk8i9687e0f42P1")
    req.Header.Set("Content-Type", "application/json")
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    fmt.Println("Response Status:", resp.Status)
    return nil
    
}