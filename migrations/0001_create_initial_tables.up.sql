-- 0001_create_initial_tables.up.sql

-- Enable UUID generation (for unique identifiers)
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create a custom type for report type (using ENUM for simplicity)
CREATE TYPE report_type AS ENUM ('radiology', 'pathology', 'lab');

-- Create a custom type for content type (using ENUM for simplicity)
CREATE TYPE content_type AS ENUM ('image', 'report', 'labtest');

-- Create the 'patientsession' table (combines Patient and Link concepts)
CREATE TABLE patientsession (
    session_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- Unique session identifier (UUID)
    access_link VARCHAR(255) UNIQUE NOT NULL, -- Unique, time-limited access link
    expiration_timestamp TIMESTAMP WITH TIME ZONE NOT NULL, -- Link expiration timestamp
    used BOOLEAN NOT NULL DEFAULT FALSE, -- Flag: has the link been used?
    patient_data JSONB,  -- Optional: patient-provided data (structured form input)
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now(), --Created at
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now() -- Updated at
);

-- Create the 'studies' table to store Study data
CREATE TABLE studies (
    id UUID PRIMARY KEY,
    patient_id UUID NOT NULL REFERENCES patientsession(session_id) ON DELETE CASCADE,
    study_instance_uid TEXT NOT NULL,
    study_data JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Create the 'images' table to store Image data
CREATE TABLE images (
    id UUID PRIMARY KEY,
    study_id UUID NOT NULL REFERENCES studies(id) ON DELETE CASCADE,
    file_path VARCHAR(255) NOT NULL,
    series_instance_uid TEXT NOT NULL,
    sop_instance_uid TEXT NOT NULL,
    image_type VARCHAR(255) NOT NULL,
    content_data JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Create the 'uploadedcontent' table (combines Study, Image, Report, Finding concepts)
CREATE TABLE uploadedcontent (
    content_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- Unique identifier for the content
    session_id UUID NOT NULL REFERENCES patientsession(session_id) ON DELETE CASCADE, -- FK to patientsession
    content_type content_type NOT NULL, -- Type of content (image, report, labtest)
    file_path VARCHAR(255) NOT NULL,  -- Path to encrypted, TEMPORARY storage
    study_data JSONB,                -- Extracted DICOM metadata (if applicable)
    content_data JSONB,              -- Extracted data (text, image metadata)
    findings JSONB,                  -- Extracted findings
    nodules JSONB,                     -- Nodule data (if applicable)
   created_at timestamp with time zone NOT NULL DEFAULT now(),
   updated_at timestamp with time zone NOT NULL DEFAULT now()
);

-- Create the 'analysisresult' table (combines Diagnosis, Stage, TreatmentRecommendation)
CREATE TABLE analysisresult (
    result_id UUID PRIMARY KEY DEFAULT gen_random_uuid(), -- Unique identifier for result
    session_id UUID NOT NULL REFERENCES patientsession(session_id) ON DELETE CASCADE, -- FK
    diagnosis JSONB,  -- Preliminary diagnosis
    stage JSONB,        -- Preliminary staging
    treatment_recommendations JSONB, -- Potential treatment options
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);

-- Create 'reports' table
CREATE TABLE reports (
    id UUID PRIMARY KEY,
    patient_id UUID NOT NULL REFERENCES patientsession(session_id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    report_type report_type NOT NULL,
    report_text TEXT,
    filepath VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Create 'findings' table - UPDATED NODULE TABLE DEFINITION
CREATE TABLE findings (
    finding_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    file_id UUID NOT NULL, --  can be linked to different file types(report, image, labtest)
    finding_type VARCHAR(255) NOT NULL,
    description TEXT NOT NULL, -- Mapped to Location in models.Nodule
    image_coordinates FLOAT [], -- Array of floats for image coordinates - Mapped to Size (first element) in models.Nodule - VERIFY!
    source VARCHAR(255) NOT NULL, -- Mapped to Shape in models.Nodule - VERIFY!
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);

-- Create 'diagnosis' table
CREATE TABLE diagnosis (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    result_id UUID NOT NULL REFERENCES analysisresult(result_id) ON DELETE CASCADE,
    session_id UUID NOT NULL REFERENCES patientsession(session_id) ON DELETE CASCADE,
    diagnosis_text TEXT,
    confidence VARCHAR(255),
    justification TEXT,
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);

-- Create 'stages' table
CREATE TABLE stages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    result_id UUID NOT NULL REFERENCES analysisresult(result_id) ON DELETE CASCADE,
    session_id UUID NOT NULL REFERENCES patientsession(session_id) ON DELETE CASCADE,
    t VARCHAR(255),
    n VARCHAR(255),
    m VARCHAR(255),
    confidence VARCHAR(255),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);

-- Create 'treatmentrecommendations' table
CREATE TABLE treatmentrecommendations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    result_id UUID NOT NULL REFERENCES analysisresult(result_id) ON DELETE CASCADE,
    session_id UUID NOT NULL REFERENCES patientsession(session_id) ON DELETE CASCADE,
    diagnosis_id UUID NOT NULL REFERENCES diagnosis(id) ON DELETE CASCADE,
    treatment_option TEXT,
    rationale TEXT,
    benefits TEXT,
    risks TEXT,
    side_effects TEXT,
    confidence VARCHAR(255),
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);


-- Create 'externalresource' table (for links to NCI, ACS, etc.)
CREATE TABLE externalresource (
    resource_id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    url VARCHAR(255) NOT NULL,
    description TEXT
);

-- Create a linking table for analysis results and external resources
CREATE TABLE analysisresultexternalresource (
    result_id UUID NOT NULL REFERENCES analysisresult(result_id) ON DELETE CASCADE,
    resource_id INTEGER NOT NULL REFERENCES externalresource(resource_id) ON DELETE CASCADE,
    PRIMARY KEY (result_id, resource_id) -- Composite primary key
);

-- Create the 'auditlog' table (for security and auditing)
CREATE TABLE auditlog (
    log_id BIGSERIAL PRIMARY KEY,  -- Auto-incrementing log ID
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP, -- Event timestamp
    action VARCHAR(255) NOT NULL,  -- Action performed (e.g., "file upload")
    details JSONB,                 -- Additional details (JSON)
    session_id UUID REFERENCES patientsession(session_id) ON DELETE SET NULL, -- FK (optional)
    content_id UUID REFERENCES uploadedcontent(content_id) ON DELETE SET NULL, -- FK (optional)
    result_id UUID REFERENCES analysisresult(result_id) ON DELETE SET NULL,    -- FK (optional)
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);

-- Create 'prompts' table to store Gemini 2.0 API prompts
CREATE TABLE prompts (
	prompt_id       UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	description     TEXT, -- purpose prompt
	template        TEXT NOT NULL, -- store prompt
	input_variables TEXT, -- input name from gemini input model
	output_format   TEXT, --  json, csv
	version         VARCHAR(255) NOT NULL,
	author          VARCHAR(255),
	status          VARCHAR(255) NOT NULL, --  draft, active, inactive
	approval_status VARCHAR(255), -- pending review, approved, rejected
    created_at timestamp with time zone NOT NULL DEFAULT now(),
    updated_at timestamp with time zone NOT NULL DEFAULT now()
);


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