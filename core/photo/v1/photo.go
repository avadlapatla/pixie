package photo

// PhotoUploaded represents a photo uploaded event
type PhotoUploaded struct {
	Id        string `json:"id"`
	Filename  string `json:"filename"`
	Mime      string `json:"mime"`
	S3Key     string `json:"s3_key"`
	CreatedAt string `json:"created_at"`
}

// PhotoDeleted represents a photo deleted event
type PhotoDeleted struct {
	Id        string `json:"id"`
	DeletedAt string `json:"deleted_at"`
}
