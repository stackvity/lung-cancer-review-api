-- 0001_create_initial_tables.down.sql

-- Drop indexes (in reverse order of creation)
DROP INDEX IF EXISTS idx_patientsession_id;
DROP INDEX IF EXISTS idx_patientsession_link;
DROP INDEX IF EXISTS idx_patientsession_exp;

DROP INDEX IF EXISTS idx_uploadedcontent_id;
DROP INDEX IF EXISTS idx_uploadedcontent_session;
DROP INDEX IF EXISTS idx_uploadedcontent_type;
DROP INDEX IF EXISTS idx_uploadedcontent_study;
DROP INDEX IF EXISTS idx_uploadedcontent_content;
DROP INDEX IF EXISTS idx_uploadedcontent_findings;
DROP INDEX IF EXISTS idx_uploadedcontent_nodules;

DROP INDEX IF EXISTS idx_analysisresult_id;
DROP INDEX IF EXISTS idx_analysisresult_session;
DROP INDEX IF EXISTS idx_analysisresult_diagnosis;
DROP INDEX IF EXISTS idx_analysisresult_stage;
DROP INDEX IF EXISTS idx_analysisresult_treatment;

DROP INDEX IF EXISTS idx_auditlog_id;
DROP INDEX IF EXISTS idx_auditlog_timestamp;
DROP INDEX IF EXISTS idx_auditlog_action;
DROP INDEX IF EXISTS idx_auditlog_details;
DROP INDEX IF EXISTS idx_auditlog_session;
DROP INDEX IF EXISTS idx_auditlog_content;
DROP INDEX IF EXISTS idx_auditlog_result;
DROP INDEX IF EXISTS idx_studies_patient_id;
DROP INDEX IF EXISTS idx_images_study_id;
DROP INDEX IF EXISTS idx_reports_patient_id;
DROP INDEX IF EXISTS idx_findings_file_id;
DROP INDEX IF EXISTS idx_diagnosis_result_id;
DROP INDEX IF EXISTS idx_stages_result_id;
DROP INDEX IF EXISTS idx_treatmentrecommendations_result_id;


-- Drop tables (in reverse order of creation to avoid dependency issues)
DROP TABLE IF EXISTS treatmentrecommendations;
DROP TABLE IF EXISTS stages;
DROP TABLE IF EXISTS diagnosis;
DROP TABLE IF EXISTS findings;
DROP TABLE IF EXISTS reports;
DROP TABLE IF EXISTS auditlog;
DROP TABLE IF EXISTS analysisresultexternalresource;
DROP TABLE IF EXISTS externalresource;
DROP TABLE IF EXISTS analysisresult;
DROP TABLE IF EXISTS uploadedcontent;
DROP TABLE IF EXISTS images;
DROP TABLE IF EXISTS patientsession;


-- Drop custom types
DROP TYPE IF EXISTS report_type;
DROP TYPE IF EXISTS content_type;