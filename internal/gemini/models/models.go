// internal/gemini/models/models.go
package models

import "github.com/google/uuid"

// NoduleDetectionInput represents the input for nodule detection for the Gemini API.
// @Description Input data structure for the Nodule Detection endpoint of the Gemini API.
type NoduleDetectionInput struct {
	ImageData           []byte  `json:"imageData" validate:"required" example:"base64 encoded image data" format:"byte" description:"Raw image data as bytes, required for the API call. Expected format depends on the Gemini API (e.g., JPEG, PNG, DICOM)."`                                                                 // ImageData (bytes):  Raw image data as bytes, required for the API call. Expected format depends on the Gemini API (e.g., JPEG, PNG, DICOM).
	ImageType           string  `json:"imageType" validate:"required" example:"image/jpeg" description:"ImageType (string):  MIME type or format of the image data (e.g., 'image/jpeg', 'application/dicom'), required for API to interpret the data correctly."`                                                              // ImageType (string):  MIME type or format of the image data (e.g., "image/jpeg", "application/dicom"), required for API to interpret the data correctly.
	Prompt              string  `json:"prompt" validate:"required" example:"Identify potential lung nodules in this image." description:"Prompt (string): User-defined or system-generated prompt to guide Gemini API's nodule detection. Should be carefully engineered for optimal results, required for task instruction."` // Prompt (string): User-defined or system-generated prompt to guide Gemini API's nodule detection.  Should be carefully engineered for optimal results, required for task instruction.
	MaxResults          int32   `json:"maxResults,omitempty" example:"5" description:"MaxResults (int32, optional): Maximum number of nodules to return in the response. Optional parameter to control API response size."`                                                                                                    // MaxResults (int32, optional): Maximum number of nodules to return in the response. Optional parameter to control API response size.
	ConfidenceThreshold float32 `json:"confidenceThreshold,omitempty" example:"0.8" description:"ConfidenceThreshold (float32, optional): Minimum confidence score for a nodule to be included in the output. Filters out low-confidence detections. Range 0.0-1.0."`                                                          // ConfidenceThreshold (float32, optional): Minimum confidence score for a nodule to be included in the output.  Filters out low-confidence detections. Range 0.0-1.0.
}

// NoduleInfo represents information about a detected nodule.
// @Description Data structure representing a single nodule detected by the Gemini API.
type NoduleInfo struct {
	Location string  `json:"location" example:"Right Upper Lobe" description:"Location (string): Anatomical location of the nodule (e.g., 'Right Upper Lobe', 'Lingula'), as identified by Gemini API."` // Location (string): Anatomical location of the nodule (e.g., "Right Upper Lobe", "Lingula"), as identified by Gemini API.
	Size     float64 `json:"size" example:"12.5" description:"Size (float64): Size of the nodule, typically in millimeters (mm), as measured by Gemini API."`                                            // Size (float64): Size of the nodule, typically in millimeters (mm), as measured by Gemini API.
	Shape    string  `json:"shape" example:"irregular" description:"Shape (string): Descriptive shape of the nodule (e.g., 'round', 'oval', 'irregular'), as characterized by Gemini API."`              // Shape (string): Descriptive shape of the nodule (e.g., "round", "oval", "irregular"), as characterized by Gemini API.
	// Add other relevant nodule characteristics as provided by Gemini API response, e.g.,
	Spiculation   string  `json:"spiculation,omitempty" example:"present" description:"Spiculation (string, optional): Spiculation characteristic of the nodule (e.g., 'present', 'absent', 'mild', 'marked'). Presence and nature of spicules radiating from the nodule."` // Spiculation (string):  Spiculation characteristic of the nodule (e.g., "present", "absent", "mild", "marked").  Presence and nature of spicules radiating from the nodule.
	Calcification string  `json:"calcification,omitempty" example:"benign" description:"Calcification (string, optional): Calcification type within the nodule (e.g., 'benign', 'malignant', 'none', 'present'). Type and pattern of calcification within the nodule."`     // Calcification (string): Calcification type within the nodule (e.g., "benign", "malignant", "none", "present"). Type and pattern of calcification within the nodule.
	Density       string  `json:"density,omitempty" example:"solid" description:"Density (string, optional): Density of the nodule (e.g., 'solid', 'part-solid', 'ground-glass'). Radiological density of the nodule, important for characterization."`                     // Density (string):  Density of the nodule (e.g., "solid", "part-solid", "ground-glass").  Radiological density of the nodule, important for characterization.
	Confidence    float64 `json:"confidence,omitempty" example:"0.95" description:"Confidence (float64): Confidence score for nodule detection (0.0-1.0 range or as provided by Gemini API). AI's confidence level in the nodule detection."`                               // Confidence (float64): Confidence score for nodule detection (0.0-1.0 range or as provided by Gemini API).  AI's confidence level in the nodule detection.
}

// NoduleDetectionOutput represents the output of nodule detection.
// @Description Output data structure for the Nodule Detection endpoint of the Gemini API.
type NoduleDetectionOutput struct {
	Nodules     []NoduleInfo `json:"nodules" description:"Nodules ([]NoduleInfo): Array of NoduleInfo structs, each representing a detected nodule. Can be empty if no nodules are detected."`                                                                                  // Nodules ([]NoduleInfo): Array of NoduleInfo structs, each representing a detected nodule. Can be empty if no nodules are detected.
	RawResponse string       `json:"rawResponse" description:"RawResponse (string): Stores the raw, unparsed JSON response from the Gemini API. Useful for debugging, error analysis, and auditing API interactions. Consider scrubbing for PHI before logging in production."` // RawResponse (string): Stores the raw, unparsed JSON response from the Gemini API.  Useful for debugging, error analysis, and auditing API interactions.  Consider scrubbing for PHI before logging in production.
	Error       string       `json:"error,omitempty" example:"API call failed due to timeout" description:"Error (string, optional): Error message from the Gemini API, if the request failed. Empty string if the request was successful."`                                    // Error (string, optional): Error message from the Gemini API, if the request failed.  Empty string if the request was successful.
}

// PathologyReportAnalysisInput is a placeholder for the input to the pathology report analysis.
// @Description Input data structure for the Pathology Report Analysis endpoint of the Gemini API.
type PathologyReportAnalysisInput struct {
	ReportText   string `json:"reportText" validate:"required" example:"Microscopic examination..." description:"ReportText (string): The full text content of the pathology report, extracted via OCR. Required for analysis."`                                                                                        // ReportText (string):  The full text content of the pathology report, extracted via OCR. Required for analysis.
	Prompt       string `json:"prompt" validate:"required" example:"Extract key findings from this pathology report." description:"Prompt (string): User-defined or system-generated prompt to guide Gemini API's pathology report analysis. Should be tailored for pathology reports, required for task instruction."` // Prompt (string): User-defined or system-generated prompt to guide Gemini API's pathology report analysis.  Should be tailored for pathology reports, required for task instruction.
	ModelVersion string `json:"modelVersion,omitempty" example:"gemini-pro-latest" description:"ModelVersion (string, optional): Specify the Gemini API model version to use for analysis. Allows for model version control if needed."`                                                                                // ModelVersion (string, optional):  Specify the Gemini API model version to use for analysis. Allows for model version control if needed.
	MaxTokens    int32  `json:"maxTokens,omitempty" example:"1024" description:"MaxTokens (int32, optional): Limits the maximum number of tokens in the API response, useful for controlling cost and response size."`                                                                                                  // MaxTokens (int32, optional): Limits the maximum number of tokens in the API response, useful for controlling cost and response size.
}

// Finding represents a single finding extracted from the pathology report.
// @Description Data structure representing a single finding extracted from a pathology report by the Gemini API.
type Finding struct {
	Description string  `json:"description" example:"Invasive adenocarcinoma" description:"Description (string): Textual description of the finding, as extracted by Gemini API. Should be patient-friendly and concise."` // Description (string): Textual description of the finding, as extracted by Gemini API.  Should be patient-friendly and concise.
	Type        string  `json:"type" example:"tumor type" description:"Type (string): Type or category of the finding (e.g., 'tumor type', 'biomarker', 'grade'). Helps categorize findings."`                             // Type (string): Type or category of the finding (e.g., "tumor type", "biomarker", "grade").  Helps categorize findings.
	Confidence  float64 `json:"confidence,omitempty" example:"0.98" description:"Confidence (float64, optional): Confidence score for the finding extraction (0.0-1.0)."`                                                  // Confidence (float64, optional):  Confidence score for the finding extraction (0.0-1.0).
	Relevance   string  `json:"relevance,omitempty" example:"high" description:"Relevance (string, optional): Relevance of the finding to lung cancer diagnosis/staging (e.g., 'high', 'medium', 'low')."`                 // Relevance (string, optional):  Relevance of the finding to lung cancer diagnosis/staging (e.g., "high", "medium", "low").
}

// PathologyReportAnalysisOutput is a placeholder for the output.
// @Description Output data structure for the Pathology Report Analysis endpoint of the Gemini API.
type PathologyReportAnalysisOutput struct {
	Findings    []Finding `json:"findings" description:"Findings ([]Finding): Array of Finding structs, each representing a key finding from the pathology report."`                   // Findings ([]Finding): Array of Finding structs, each representing a key finding from the pathology report.
	RawResponse string    `json:"rawResponse" description:"RawResponse (string): Stores the raw JSON response from the Gemini API for debugging/auditing."`                            // RawResponse (string): Stores the raw JSON response from the Gemini API for debugging/auditing.
	Error       string    `json:"error,omitempty" example:"API call quota exceeded" description:"Error (string, optional): Error message from the Gemini API, if the request failed."` // Error (string, optional): Error message from the Gemini API, if the request failed.
}

// InformationExtractionInput is a placeholder for input to general information extraction.
// @Description Input data structure for the Information Extraction endpoint of the Gemini API. Used for general-purpose information extraction from text.
type InformationExtractionInput struct {
	ReportText string `json:"reportText" validate:"required" example:"Patient presented with cough and chest pain..." description:"ReportText (string): Text content of the report for information extraction. Required."`             // ReportText (string): Text content of the report for information extraction. Required.
	Prompt     string `json:"prompt" validate:"required" example:"Extract all symptoms mentioned in this report." description:"Prompt (string): Prompt to guide Gemini API's information extraction. Required."`                       // Prompt (string): Prompt to guide Gemini API's information extraction. Required.
	TaskType   string `json:"taskType,omitempty" example:"symptoms" description:"TaskType (string, optional): Specify the type of information to extract (e.g., 'symptoms', 'treatment'). Allows for task-specific prompting."`        // TaskType (string, optional):  Specific the type of information to extract (e.g., "symptoms", "treatment").  Allows for task-specific prompting.
	Context    string `json:"context,omitempty" example:"Patient is being evaluated for lung cancer." description:"Context (string, optional): Additional context to guide information extraction (e.g., 'patient medical history')."` // Context (string, optional):  Additional context to guide information extraction (e.g., "patient medical history").
}

// InformationExtractionOutput is a placeholder for the output of information extraction.
// @Description  Output data structure for the Information Extraction endpoint of the Gemini API.
type InformationExtractionOutput struct {
	Findings    []Finding `json:"findings" description:"Findings ([]Finding): Array of Finding structs, representing extracted pieces of information."`                // Findings ([]Finding): Array of Finding structs, representing extracted pieces of information.
	RawResponse string    `json:"rawResponse" description:"RawResponse (string): Raw JSON response from Gemini API."`                                                  // RawResponse (string): Raw JSON response from Gemini API.
	Error       string    `json:"error,omitempty" example:"Gemini API service unavailable" description:"Error (string, optional): Error message from the Gemini API."` // Error (string, optional): Error message from the Gemini API.
}

// DiagnosisInput represents the input for preliminary diagnosis generation.
// @Description Input data structure for the Preliminary Diagnosis generation endpoint of the Gemini API.
type DiagnosisInput struct {
	PatientID       uuid.UUID `json:"patientId" validate:"required,uuid4" example:"a1b2c3d4-e5f6-4789-9012-34567890abcd" description:"PatientID (UUID): Unique identifier for the patient session, UUID format, required and validated as UUIDv4."`          // PatientID (UUID):  Unique identifier for the patient session, UUID format, required and validated as UUIDv4.
	Prompt          string    `json:"prompt" validate:"required" example:"Generate a preliminary lung cancer diagnosis..." description:"Prompt (string): Prompt for guiding Gemini API's diagnosis generation, required."`                                   // Prompt (string): Prompt for guiding Gemini API's diagnosis generation, required.
	MedicalHistory  string    `json:"medicalHistory,omitempty" example:"Patient is a 58-year-old former smoker..." description:"MedicalHistory (string): Patient's medical history as text. Optional, but improves diagnostic accuracy."`                    // MedicalHistory (string): Patient's medical history as text.  Optional, but improves diagnostic accuracy.
	Symptoms        string    `json:"symptoms,omitempty" example:"Persistent cough, shortness of breath" description:"Symptoms (string): Patient's reported symptoms as text. Optional, but improves diagnostic accuracy."`                                  // Symptoms (string):  Patient's reported symptoms as text. Optional, but improves diagnostic accuracy.
	FindingsSummary string    `json:"findingsSummary,omitempty" example:"CT scan shows a 2cm nodule in the right upper lobe..." description:"FindingsSummary (string): Summary of findings extracted from reports. Optional, but provides crucial context."` // FindingsSummary (string): Summary of findings extracted from reports. Optional, but provides crucial context.
	ReportText      string    `json:"reportText,omitempty" example:"..." description:"ReportText (string): Full report text for context. Optional, but can improve diagnostic accuracy."`                                                                    // ReportText (string): Full report text for context.  Optional, but can improve diagnostic accuracy.
}

// DiagnosisOutput represents the output of preliminary diagnosis generation.
// @Description Output data structure for the Preliminary Diagnosis generation endpoint of the Gemini API.
type DiagnosisOutput struct {
	Diagnosis     string `json:"diagnosis" description:"Diagnosis (string): Preliminary diagnosis text, generated by Gemini API, patient-friendly language."`                                                // Diagnosis (string): Preliminary diagnosis text, generated by Gemini API, patient-friendly language.
	Confidence    string `json:"confidence" example:"Moderate" description:"Confidence (string): Confidence level of the preliminary diagnosis, as provided by Gemini API (e.g., 'High', 'Medium', 'Low')."` // Confidence (string): Confidence level of the preliminary diagnosis, as provided by Gemini API (e.g., "High", "Medium", "Low").
	Justification string `json:"justification" description:"Justification (string): Explanation or justification for the diagnosis, generated by Gemini API, plain language."`                               // Justification (string): Explanation or justification for the diagnosis, generated by Gemini API, plain language.
	RawResponse   string `json:"rawResponse" description:"RawResponse (string): Raw JSON response from Gemini API."`                                                                                         // RawResponse (string): Raw JSON response from Gemini API.
	Error         string `json:"error,omitempty" example:"Model service error" description:"Error (string, optional): Error message from the Gemini API."`                                                   // Error (string, optional): Error message from the Gemini API.
}

// StagingInput represents the input for staging information retrieval.
// @Description Input data structure for the Staging Information retrieval endpoint of the Gemini API, used to get TNM staging.
type StagingInput struct {
	PatientID uuid.UUID `json:"patientId" validate:"required,uuid4" example:"a1b2c3d4-e5f6-4789-9012-34567890abcd" description:"PatientID (UUID): Patient session identifier, required, UUIDv4 format."`  // PatientID (UUID): Patient session identifier, required, UUIDv4 format.
	Prompt    string    `json:"prompt" validate:"required" example:"Determine TNM staging for this lung cancer case." description:"Prompt (string): Prompt for staging information retrieval, required."` // Prompt (string): Prompt for staging information retrieval, required.
	T         string    `json:"T" example:"T2a" description:"T (string): TNM staging - Tumor component (if available from prior analysis or input)."`                                                     // T (string):  T stage component (if available from prior analysis or input).
	N         string    `json:"N" example:"N1" description:"N (string): TNM staging - Node component."`                                                                                                   // N (string):  N stage component.
	M         string    `json:"M" example:"M0" description:"M (string): TNM staging - Metastasis component."`                                                                                             // M (string):  M stage component.
	// Add fields for relevant patient data for staging (if needed) // Consider adding fields for structured patient data relevant for staging, e.g.,
	NoduleSize           string `json:"noduleSize,omitempty" example:"2 cm" description:"NoduleSize (string): Size of the nodule (if detected), to help Gemini determine T stage."`                                                     // NoduleSize (string): Size of the nodule (if detected), to help Gemini determine T stage.
	LymphNodeInvolvement string `json:"lymphNodeInvolvement,omitempty" example:"Mediastinal lymph nodes enlarged" description:"LymphNodeInvolvement (string): Information on lymph node involvement from reports/images, for N stage."` // LymphNodeInvolvement (string): Information on lymph node involvement from reports/images, for N stage.
	MetastasisPresent    string `json:"metastasisPresent,omitempty" example:"No distant metastasis" description:"MetastasisPresent (string): Information on distant metastasis, for M stage."`                                          // MetastasisPresent (string):  Information on distant metastasis, for M stage.
	FindingsSummary      string `json:"findingsSummary,omitempty" example:"Findings are suggestive of early-stage lung cancer." description:"FindingsSummary (string): Summary of findings, to provide context for staging."`           // FindingsSummary (string): Summary of findings, to provide context for staging.
}

// StagingOutput represents the output for staging information.
// @Description Output data structure for the Staging Information retrieval endpoint of the Gemini API.
type StagingOutput struct {
	T           string `json:"T" example:"T2a" description:"T (string): TNM staging - Tumor component, extracted from Gemini API."`                                                              // T (string): TNM staging - Tumor component, extracted from Gemini API.
	N           string `json:"N" example:"N1" description:"N (string): TNM staging - Node component."`                                                                                           // N (string): TNM staging - Node component.
	M           string `json:"M" example:"M0" description:"M (string): TNM staging - Metastasis component."`                                                                                     // M (string): TNM staging - Metastasis component.
	StageValue  string `json:"stageValue" example:"Stage IIB" description:"StageValue (string): Overall TNM stage (e.g., 'Stage IA', 'Stage IIIB'), derived by Gemini API or rule-based logic."` // StageValue (string): Overall TNM stage (e.g., "Stage IA", "Stage IIIB"), derived by Gemini API or rule-based logic.
	Confidence  string `json:"confidence" example:"High" description:"Confidence (string): Confidence level for the staging information, as provided by Gemini API."`                            // Confidence (string): Confidence level for the staging information, as provided by Gemini API.
	RawResponse string `json:"rawResponse" description:"RawResponse (string): Raw JSON response from Gemini API."`                                                                               // RawResponse (string): Raw JSON response from Gemini API.
	Error       string `json:"error,omitempty" example:"Model API error" description:"Error (string, optional): Error message from the Gemini API."`                                             // Error (string, optional): Error message from the Gemini API.
}

// TreatmentRecommendationInput represents the input for treatment recommendation suggestions.
// @Description Input data structure for the Treatment Recommendation Suggestion endpoint of the Gemini API.
type TreatmentRecommendationInput struct {
	PatientID uuid.UUID `json:"patientId" validate:"required,uuid4" example:"a1b2c3d4-e5f6-4789-9012-34567890abcd" description:"PatientID (UUID): Patient session identifier, required, UUIDv4 format."`                // PatientID (UUID): Patient session identifier, required, UUIDv4 format.
	Prompt    string    `json:"prompt" validate:"required" example:"Suggest treatment options for Stage IIB lung cancer." description:"Prompt (string): Prompt to guide Gemini API's treatment suggestions, required."` // Prompt (string): Prompt to guide Gemini API's treatment suggestions, required.
	Diagnosis string    `json:"diagnosis" example:"Non-small cell lung cancer, likely adenocarcinoma" description:"Diagnosis (string): Preliminary diagnosis text (if available, to provide context to Gemini)."`       // Diagnosis (string): Preliminary diagnosis text (if available, to provide context to Gemini).
	Stage     string    `json:"stage" example:"Stage IIB" description:"Stage (string): Preliminary staging information (if available)."`                                                                                // Stage (string): Preliminary staging information (if available).
	// Add fields for other factors influencing treatment (e.g., patient comorbidities, preferences) // Consider adding fields for patient-specific factors, e.g.,
	PatientAge         int    `json:"patientAge" validate:"omitempty,min=18,max=120" example:"58" description:"PatientAge (int): Patient's age, which can influence treatment decisions. Optional, validated for realistic age range if provided."`                                      // PatientAge (int): Patient's age, which can influence treatment decisions. Optional, validated for realistic age range if provided.
	Comorbidities      string `json:"comorbidities" example:"Hypertension, COPD" description:"Comorbidities (string): List of patient's comorbidities or other health conditions. Optional."`                                                                                            // Comorbidities (string):  List of patient's comorbidities or other health conditions. Optional.
	PatientPreferences string `json:"patientPreferences" example:"Patient prefers non-surgical options if possible." description:"PatientPreferences (string): Patient's expressed preferences or values regarding treatment (e.g., 'patient prefers non-surgical options'). Optional."` // PatientPreferences (string):  Patient's expressed preferences or values regarding treatment (e.g., "patient prefers non-surgical options"). Optional.
	Biomarkers         string `json:"biomarkers" example:"EGFR mutation positive" description:"Biomarkers (string): Key biomarker information from pathology reports (e.g., 'EGFR mutation positive'). Optional."`                                                                       // Biomarkers (string):  Key biomarker information from pathology reports (e.g., "EGFR mutation positive"). Optional.
}

// TreatmentRecommendationInfo represents a single treatment recommendation.
// @Description Data structure representing a single treatment recommendation suggested by the Gemini API.
type TreatmentRecommendationInfo struct {
	TreatmentOption  string `json:"treatmentOption" example:"Surgery" description:"TreatmentOption (string): Name of the potential treatment option (e.g., 'Surgery', 'Chemotherapy')."`                                                                                     // TreatmentOption (string): Name of the potential treatment option (e.g., "Surgery", "Chemotherapy").
	Rationale        string `json:"rationale" example:"Surgical resection is the standard of care for early-stage NSCLC." description:"Rationale (string): Justification for suggesting this treatment, plain language."`                                                    // Rationale (string): Justification for suggesting this treatment, plain language.
	Benefits         string `json:"benefits" example:"Potentially curative in early stages." description:"Benefits (string): Potential benefits of the treatment, plain language, patient-friendly."`                                                                        // Benefits (string): Potential benefits of the treatment, plain language, patient-friendly.
	Risks            string `json:"risks" example:"Bleeding, infection, pain." description:"Risks (string): Potential risks associated with the treatment, plain language, patient-friendly."`                                                                               // Risks (string): Potential risks associated with the treatment, plain language, patient-friendly.
	SideEffects      string `json:"sideEffects" example:"Fatigue, nausea." description:"SideEffects (string): Common side effects of the treatment, plain language, patient-friendly."`                                                                                      // SideEffects (string): Common side effects of the treatment, plain language, patient-friendly.
	Confidence       string `json:"confidence" example:"High" description:"Confidence (string): Confidence level for the treatment recommendation, as provided by Gemini API."`                                                                                              // Confidence (string): Confidence level for the treatment recommendation, as provided by Gemini API.
	Source           string `json:"source" example:"NCCN Guidelines" description:"Source (string): Source of the recommendation (e.g., 'NCCN Guidelines', 'Gemini API')."`                                                                                                   // Source (string): Source of the recommendation (e.g., "NCCN Guidelines", "Gemini API").
	CostEstimate     string `json:"costEstimate" example:"High" description:"CostEstimate (string, optional): Estimated cost of treatment (e.g., 'High', 'Moderate', 'Low', or a numerical range if available). Use cautiously and with disclaimers about cost variations."` // CostEstimate (string, optional):  Estimated cost of treatment (e.g., "High", "Moderate", "Low", or a numerical range if available).  Use cautiously and with disclaimers about cost variations.
	DurationEstimate string `json:"durationEstimate" example:"4-6 weeks" description:"DurationEstimate (string, optional): Estimated duration of treatment (e.g., '6 months', '4-6 weeks'). Use cautiously and with disclaimers about individual variability."`              // DurationEstimate (string, optional): Estimated duration of treatment (e.g., "6 months", "4-6 weeks").  Use cautiously and with disclaimers about individual variability.
}

// TreatmentRecommendationOutput represents the output for treatment recommendations.
// @Description Output data structure for the Treatment Recommendation Suggestion endpoint of the Gemini API.
type TreatmentRecommendationOutput struct {
	Recommendations []*TreatmentRecommendationInfo `json:"recommendations" description:"Recommendations ([]*TreatmentRecommendationInfo): Array of treatment recommendations, each with details on option, rationale, benefits, risks, etc."` // Recommendations ([]*TreatmentRecommendationInfo): Array of treatment recommendations, each with details on option, rationale, benefits, risks, etc.
	RawResponse     string                         `json:"rawResponse" description:"RawResponse (string): Raw JSON response from Gemini API."`                                                                                                // RawResponse (string): Raw JSON response from Gemini API.
	Error           string                         `json:"error,omitempty" example:"Model did not provide specific recommendations" description:"Error (string, optional): Error message from the Gemini API."`                               // Error (string, optional): Error message from the Gemini API.
}
