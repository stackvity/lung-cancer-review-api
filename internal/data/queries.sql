-- internal/data/queries.sql

-- Add your SQL queries here

-- Create a custom type for report type (using ENUM for simplicity)
CREATE TYPE report_type AS ENUM ('radiology', 'pathology', 'lab');

-- Create a custom type for content type (using ENUM for simplicity)
CREATE TYPE content_type AS ENUM ('image', 'report', 'labtest');

-- ------------- PatientSession (and Link) Queries -------------

-- CreatePatientSession: Creates a new patient session (with associated link).
-- name: CreatePatientSession :one
INSERT INTO patientsession (access_link, expiration_timestamp, used, patient_data)
VALUES ($1, $2, $3, $4)
RETURNING session_id, access_link, expiration_timestamp, used, patient_data, created_at, updated_at;

-- GetPatientSessionByLink: Retrieves a patient session by its access link.
-- name: GetPatientSessionByLink :one
SELECT session_id, access_link, expiration_timestamp, used, patient_data, created_at, updated_at FROM patientsession
WHERE access_link = $1;

-- UpdatePatientSessionUsed: Marks a patient session as used.
-- name: UpdatePatientSessionUsed :exec
UPDATE patientsession
SET used = TRUE
WHERE session_id = $1;

-- InvalidateLink: Sets a link to used (effectively invalidating it).
-- name: InvalidateLink :exec
UPDATE patientsession
SET used = TRUE
WHERE access_link = $1;

-- DeleteExpiredSessions: Deletes expired patient sessions.
-- name: DeleteExpiredSessions :exec
DELETE FROM patientsession
WHERE expiration_timestamp < NOW();

-- DeletePatientSession: Deletes a patient session by session ID.
-- name: DeletePatientSession :exec
DELETE from patientsession
WHERE session_id = $1;


-- ------------- UploadedContent Queries -------------

-- CreateUploadedContent: Inserts a new uploaded content record.
-- name: CreateUploadedContent :one
INSERT INTO uploadedcontent (session_id, content_type, file_path, study_data, content_data, findings, nodules)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING content_id, session_id, content_type, file_path, study_data, content_data, findings, nodules, created_at, updated_at;

-- GetUploadedContentByID: Retrieves uploaded content by its ID.
-- name: GetUploadedContentByID :one
SELECT content_id, session_id, content_type, file_path, study_data, content_data, findings, nodules, created_at, updated_at FROM uploadedcontent
WHERE content_id = $1;

-- ListUploadedContentBySessionID: Retrieves all uploaded content for a given session.
-- name: ListUploadedContentBySessionID :many
SELECT content_id, session_id, content_type, file_path, study_data, content_data, findings, nodules, created_at, updated_at FROM uploadedcontent
WHERE session_id = $1
ORDER BY created_at DESC;


-- DeleteUploadedContent: Deletes uploaded content by its ID.
-- name: DeleteUploadedContent :exec
DELETE FROM uploadedcontent
WHERE content_id = $1;

-- DeleteUploadedContentBySessionID: Deletes all uploaded content for a session.
-- name: DeleteUploadedContentBySessionID :exec
DELETE FROM uploadedcontent
WHERE session_id = $1;

-- ------------- AnalysisResult Queries -------------

-- CreateAnalysisResult: Inserts a new analysis result record.
-- name: CreateAnalysisResult :one
INSERT INTO analysisresult (session_id, diagnosis, stage, treatment_recommendations)
VALUES ($1, $2, $3, $4)
RETURNING result_id, session_id, diagnosis, stage, treatment_recommendations, created_at, updated_at;

-- GetAnalysisResultByID: Retrieves an analysis result by its ID.
-- name: GetAnalysisResultByID :one
SELECT result_id, session_id, diagnosis, stage, treatment_recommendations, created_at, updated_at FROM analysisresult
WHERE result_id = $1;

-- GetAnalysisResultBySessionID: Retrieves the analysis result for a given session.
-- name: GetAnalysisResultBySessionID :one
SELECT result_id, session_id, diagnosis, stage, treatment_recommendations, created_at, updated_at FROM analysisresult
WHERE session_id = $1;


-- DeleteAnalysisResult: Deletes an analysis result by its ID.
-- name: DeleteAnalysisResult :exec
DELETE FROM analysisresult
WHERE result_id = $1;

-- DeleteAnalysisResultBySessionID: Deletes the analysis result for a session.
-- name: DeleteAnalysisResultBySessionID :exec
DELETE FROM analysisresult
WHERE session_id = $1;


-- ------------- AuditLog Queries -------------

-- CreateAuditLog: Inserts a new audit log entry.
-- name: CreateAuditLog :one
INSERT INTO auditlog (session_id, content_id, result_id, action, details)
VALUES ($1, $2, $3, $4, $5)
RETURNING log_id, timestamp, action, details, session_id, content_id, result_id, created_at, updated_at;

-- GetAuditLogByID: Retrieves an audit log entry by its ID.
-- name: GetAuditLogByID :one
SELECT log_id, timestamp, action, details, session_id, content_id, result_id, created_at, updated_at FROM auditlog
WHERE log_id = $1;

-- ListAuditLogsBySessionID: Retrieves all audit log entries for a given session.
-- name: ListAuditLogsBySessionID :many
SELECT log_id, timestamp, action, details, session_id, content_id, result_id, created_at, updated_at FROM auditlog
WHERE session_id = $1
ORDER BY timestamp DESC;

-- ListAuditLogsByContentID: Retrieves all audit log entries for a given content item.
-- name: ListAuditLogsByContentID :many
SELECT log_id, timestamp, action, details, session_id, content_id, result_id, created_at, updated_at FROM auditlog
WHERE content_id = $1
ORDER BY timestamp DESC;

-- ListAuditLogsByResultID: Retrieves audit logs for a given analysis result
-- name: ListAuditLogsByResultID :many
SELECT log_id, timestamp, action, details, session_id, content_id, result_id, created_at, updated_at FROM auditlog
WHERE result_id = $1
ORDER BY timestamp DESC;

-- ListAuditLogsByAction: Retrieves all audit log entries for a specific action.
-- name: ListAuditLogsByAction :many
SELECT log_id, timestamp, action, details, session_id, content_id, result_id, created_at, updated_at FROM auditlog
WHERE action = $1
ORDER BY timestamp DESC;


-- ------------- ExternalResource Queries -------------
-- These don't interact with patient data directly, so they are less sensitive.

-- CreateExternalResource: Inserts a new external resource.
-- name: CreateExternalResource :one
INSERT INTO externalresource (name, url, description)
VALUES ($1, $2, $3)
RETURNING resource_id, name, url, description;

-- GetExternalResourceByID: Retrieves an external resource by its ID.
-- name: GetExternalResourceByID :one
SELECT resource_id, name, url, description FROM externalresource
WHERE resource_id = $1;

-- ListExternalResources: Retrieves all external resources.
-- name: ListExternalResources :many
SELECT resource_id, name, url, description FROM externalresource
ORDER BY name;

-- UpdateExternalResource: Updates an existing external resource.
-- name: UpdateExternalResource :one
UPDATE externalresource
SET name = $2, url = $3, description = $4
WHERE resource_id = $1
RETURNING resource_id, name, url, description;

-- DeleteExternalResource: Deletes an external resource by its ID.
-- name: DeleteExternalResource :exec
DELETE FROM externalresource
WHERE resource_id = $1;

-- CreateAnalysisResultExternalResource: Links an analysis result to an external resource.
-- name: CreateAnalysisResultExternalResource :exec
INSERT INTO analysisresultexternalresource (result_id, resource_id)
VALUES ($1, $2);

-- DeleteAnalysisResultExternalResource: Removes a link between an analysis result and an external resource.
-- name: DeleteAnalysisResultExternalResource :exec
DELETE FROM analysisresultexternalresource WHERE result_id = $1 and resource_id = $2;

-- ListExternalResourcesByResultID: Get all external resources associated with a given analysis result.
-- name: ListExternalResourcesByResultID :many
SELECT er.resource_id, er.name, er.url, er.description
FROM externalresource er
         INNER JOIN analysisresultexternalresource ar ON er.resource_id = ar.resource_id
WHERE ar.result_id = $1;

-- ------------- Prompt Queries -------------

-- CreatePrompt: Inserts a new prompt.
-- name: CreatePrompt :one
INSERT INTO prompts (description, template, input_variables, output_format, version, author, status, approval_status)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING prompt_id, description, template, input_variables, output_format, version, author, status, approval_status, created_at, updated_at;

-- GetPromptByID: Retrieves a prompt by its ID.
-- name: GetPromptByID :one
SELECT prompt_id, description, template, input_variables, output_format, version, author, status, approval_status, created_at, updated_at
FROM prompts
WHERE prompt_id = $1;

-- GetActivePrompt: Retrieves the currently active prompt by description.
-- name: GetActivePrompt :one
SELECT prompt_id, description, template, input_variables, output_format, version, author, status, approval_status, created_at, updated_at
FROM prompts
WHERE description = $1
AND status = 'active'
ORDER BY version DESC
LIMIT 1;

-- ListPrompts: Retrieves all prompts.
-- name: ListPrompts :many
SELECT prompt_id, description, template, input_variables, output_format, version, author, status, approval_status, created_at, updated_at
FROM prompts
ORDER BY description, version;

-- UpdatePrompt: Updates an existing prompt.
-- name: UpdatePrompt :one
UPDATE prompts
SET description     = $2,
	template        = $3,
	input_variables = $4,
	output_format   = $5,
	version         = $6,
	author          = $7,
	status          = $8,
	approval_status = $9
WHERE prompt_id = $1
RETURNING prompt_id, description, template, input_variables, output_format, version, author, status, approval_status, created_at, updated_at;

-- DeletePrompt: Deletes a prompt by its ID.
-- name: DeletePrompt :exec
DELETE
FROM prompts
WHERE prompt_id = $1;


-- ------------- Study Queries -------------

-- CreateStudy creates a new study
-- name: CreateStudy :one
INSERT INTO studies (id, patient_id, study_instance_uid, study_data)
VALUES ($1, $2, $3, $4)
RETURNING id, patient_id, study_instance_uid, study_data, created_at, updated_at;

-- GetStudyByID retrieves a study by its ID
-- name: GetStudyByID :one
SELECT id, patient_id, study_instance_uid, study_data, created_at, updated_at FROM studies
WHERE id = $1;

-- ListStudiesByPatientID retrieves all studies for a patient
-- name: ListStudiesByPatientID :many
SELECT id, patient_id, study_instance_uid, study_data, created_at, updated_at FROM studies
WHERE patient_id = $1;

-- DeleteStudy deletes a study by ID
-- name: DeleteStudy :exec
DELETE FROM studies
WHERE id = $1;

-- DeleteAllStudiesByPatientID deletes all studies for a patient
-- name: DeleteAllStudiesByPatientID :exec
DELETE FROM studies
WHERE patient_id = $1;


-- ------------- Image Queries -------------

-- CreateImage creates a new image
-- name: CreateImage :one
INSERT INTO images (id, study_id, file_path, series_instance_uid, sop_instance_uid, image_type, content_data)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, study_id, file_path, series_instance_uid, sop_instance_uid, image_type, content_data, created_at, updated_at;

-- GetImageByID retrieves a image by its ID
-- name: GetImageByID :one
SELECT id, study_id, file_path, series_instance_uid, sop_instance_uid, image_type, content_data, created_at, updated_at FROM images
WHERE id = $1;

-- GetImageByStudyID retrieves all images for a study
-- name: GetImageByStudyID :many
SELECT id, study_id, file_path, series_instance_uid, sop_instance_uid, image_type, content_data, created_at, updated_at FROM images
WHERE study_id = $1;

-- ListImagesByPatientID retrieves all images for a patient
-- name: ListImagesByPatientID :many
SELECT images.id, images.study_id, images.file_path, images.series_instance_uid, images.sop_instance_uid, images.image_type, images.content_data, images.created_at, images.updated_at
FROM images
INNER JOIN studies ON images.study_id = studies.id
WHERE studies.patient_id = $1;


-- DeleteImage deletes a image by ID
-- name: DeleteImage :exec
DELETE FROM images
WHERE id = $1;

-- DeleteAllImagesByPatientID deletes all image records associated with a given patient ID.
-- name: DeleteAllImagesByPatientID :exec
DELETE FROM images
WHERE study_id IN (SELECT id FROM studies WHERE patient_id = $1);


-- ------------- Report Queries -------------

-- CreateReport inserts a new report record.
-- name: CreateReport :one
INSERT INTO reports (id, patient_id, filename, report_type, report_text, filepath)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, patient_id, filename, report_type, report_text, filepath, created_at, updated_at;

-- GetReportByID retrieves a report by its ID.
-- name: GetReportByID :one
SELECT id, patient_id, filename, report_type, report_text, filepath, created_at, updated_at
FROM reports
WHERE id = $1;

-- GetReportByPatientID retrieves all reports for a patient
-- name: GetReportByPatientID :many
SELECT id, patient_id, filename, report_type, report_text, filepath, created_at, updated_at
FROM reports
WHERE patient_id = $1;

-- DeleteReport deletes a report by ID
-- name: DeleteReport :exec
DELETE FROM reports
WHERE id = $1;

-- DeleteAllReportsByPatientID deletes all reports for a patient
-- name: DeleteAllReportsByPatientID :exec
DELETE FROM reports
WHERE patient_id = $1;

-- CreateFinding inserts a new finding record.
-- name: CreateFinding :one
INSERT INTO findings (finding_id, file_id, finding_type, description, image_coordinates, source)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING finding_id, file_id, finding_type, description, image_coordinates, source, created_at, updated_at;

-- GetNoduleByID retrieves a nodule by its ID.
-- name: GetNoduleByID :one
SELECT finding_id, file_id, description, image_coordinates, source FROM findings WHERE finding_id = $1;


-- ------------- Diagnosis Queries -------------

-- CreateDiagnosis inserts a new diagnosis record.
-- name: CreateDiagnosis :one
INSERT INTO diagnosis (result_id, session_id, diagnosis_text, confidence, justification)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, result_id, session_id, diagnosis_text, confidence, justification, created_at, updated_at;

-- GetDiagnosisByID retrieves a diagnosis by its ID.
-- name: GetDiagnosisByID :one
SELECT id, result_id, session_id, diagnosis_text, confidence, justification, created_at, updated_at FROM diagnosis
WHERE id = $1;

-- ------------- Stage Queries -------------

-- CreateStaging inserts a new staging record.
-- name: CreateStaging :one
INSERT INTO stages (result_id, session_id, t, n, m, confidence)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id, result_id, session_id, t, n, m, confidence, created_at, updated_at;

-- GetStageByID retrieves a staging record by its ID.
-- name: GetStageByID :one
SELECT id, result_id, session_id, t, n, m, confidence, created_at, updated_at
FROM stages
WHERE id = $1;


-- ------------- TreatmentRecommendation Queries -------------

-- CreateTreatmentRecommendation inserts a new treatment recommendation record.
-- name: CreateTreatmentRecommendation :one
INSERT INTO treatmentrecommendations (result_id, session_id, diagnosis_id, treatment_option, rationale, benefits, risks, side_effects, confidence)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING id, result_id, session_id, diagnosis_id, treatment_option, rationale, benefits, risks, side_effects, confidence, created_at, updated_at;

-- GetTreatmentRecommendationByID retrieves a treatment recommendation record by its ID.
-- name: GetTreatmentRecommendationByID :one
SELECT id, result_id, session_id, diagnosis_id, treatment_option, rationale, benefits, risks, side_effects, confidence, created_at, updated_at
FROM treatmentrecommendations
WHERE id = $1;


-- Add indexes for performance (on frequently queried columns)
CREATE INDEX idx_patientsession_id ON patientsession(session_id);
CREATE INDEX idx_patientsession_link ON patientsession(access_link);
CREATE INDEX idx_patientsession_exp ON patientsession(expiration_timestamp);

CREATE INDEX idx_uploadedcontent_id ON uploadedcontent(content_id);
CREATE INDEX idx_uploadedcontent_session ON uploadedcontent(session_id);
CREATE INDEX idx_uploadedcontent_type ON uploadedcontent(content_type);
CREATE INDEX idx_uploadedcontent_study ON uploadedcontent USING GIN (study_data);  -- GIN index
CREATE INDEX idx_uploadedcontent_content ON uploadedcontent USING GIN (content_data); -- GIN index
CREATE INDEX idx_uploadedcontent_findings ON uploadedcontent USING GIN (findings);    -- GIN index for JSONB
CREATE INDEX idx_uploadedcontent_nodules ON uploadedcontent USING GIN (nodules);      -- GIN index for JSONB

CREATE INDEX idx_analysisresult_id ON analysisresult(result_id);
CREATE INDEX idx_analysisresult_session ON analysisresult(session_id);
CREATE INDEX idx_analysisresult_diagnosis ON analysisresult USING GIN (diagnosis); -- GIN index
CREATE INDEX idx_analysisresult_stage ON analysisresult USING GIN (stage);             -- GIN index
CREATE INDEX idx_analysisresult_treatment ON analysisresult USING GIN (treatment_recommendations); -- GIN

CREATE INDEX idx_auditlog_id ON auditlog(log_id);
CREATE INDEX idx_auditlog_timestamp ON auditlog(timestamp);
CREATE INDEX idx_auditlog_action ON auditlog(action);
CREATE INDEX idx_auditlog_details ON auditlog USING GIN (details);  -- GIN index
CREATE INDEX idx_auditlog_session ON auditlog(session_id);
CREATE INDEX idx_auditlog_content ON auditlog(content_id);
CREATE INDEX idx_auditlog_result ON auditlog(result_id);
CREATE INDEX idx_studies_patient_id ON studies(patient_id);
CREATE INDEX idx_images_study_id ON images(study_id);
CREATE INDEX idx_reports_patient_id ON reports(patient_id);
CREATE INDEX idx_findings_file_id ON findings(file_id);
CREATE INDEX idx_diagnosis_result_id ON diagnosis(result_id);
CREATE INDEX idx_stages_result_id ON stages(result_id);
CREATE INDEX idx_treatmentrecommendations_result_id ON treatmentrecommendations(result_id);