// internal/ocr/ocr.go
package ocr

import "context"

// OCRService defines the interface for OCR functionality.
//
//	This interface outlines the contract for OCR service implementations,
//	specifying the method for extracting text from images.
//	Implementations MUST consider:
//	  - Robust Error Handling: Implement comprehensive error handling for various OCR failure scenarios.
//	    (See Error Handling section in method documentation below).  *THIS IS MANDATORY.*
//	  - Performance: Optimize for latency and throughput, especially for handling large documents or high request volumes.
//	    (Consider asynchronous processing and caching mechanisms). *CRITICAL for user experience and scalability.*
//	  - Security: Ensure secure handling of image data and API keys (if using cloud services). *ABSOLUTELY ESSENTIAL for data privacy and compliance.*
//	  - Metrics and Monitoring: Integrate metrics and logging for performance monitoring and error tracking in production. *REQUIRED for operational visibility and proactive issue resolution.*
type OCRService interface {
	// ExtractText performs Optical Character Recognition (OCR) on the provided image data.
	// It is the core method of the OCRService interface, responsible for the text extraction process.
	//
	// Implementations MUST:
	//   - Handle various image formats (JPEG, PNG, PDF, DICOM image frames, etc.) based on the imageType parameter.
	//     Implement robust format detection and processing to ensure broad compatibility.
	//   - Utilize an appropriate OCR engine or service to perform text extraction, selecting the best option based on accuracy, performance, and cost considerations.
	//   - Return the extracted text, a confidence score (if available from the OCR engine), and any errors encountered during the process.
	//   - Implement context cancellation and timeout to ensure responsiveness and prevent indefinite operations. *Context handling is REQUIRED for production readiness.*
	//
	// Parameters:
	//   - ctx context.Context: Context for cancellation and timeout.  Implementations MUST respect context deadlines and cancellations to prevent resource leaks and ensure responsiveness under load.
	//   - imageData []byte: Raw image data as bytes. The format should be compatible
	//     with the chosen OCR engine (e.g., JPEG, PNG, PDF). Implementations MUST handle different image formats and validate input data.
	//   - imageType string: MIME type or format of the image data. This helps the OCR
	//     service to correctly interpret the image data (e.g., "image/jpeg", "application/pdf", "application/dicom").  Implementations MUST use this parameter to guide OCR processing.
	//
	// Returns:
	//   - string: Extracted text from the image. On success, returns the text content
	//     obtained from performing OCR on the image. The text SHOULD be cleaned and formatted for further processing (e.g., UTF-8 encoding, normalized whitespace) to ensure consistency and usability.
	//   - float64: Confidence score of the OCR extraction. This value represents the
	//     overall confidence level of the OCR process for the entire document or image.
	//     The range and meaning of this score are implementation-specific (e.g., 0.0-1.0, percentage) and MUST be thoroughly documented by each implementation.
	//     Return 0.0 if confidence score is not available or not applicable, but document this behavior clearly.
	//   - error: An error if the OCR extraction fails.  This could be due to various reasons,
	//     such as network issues with a cloud OCR service, invalid image data,
	//     unsupported image format, or internal OCR engine errors.  Implementations MUST return `nil` on successful text extraction, even if the extracted text is empty, and a *specific, custom error type* from `internal/domain/errors.go` on failure.
	//
	// Error Handling: *ROBUST ERROR HANDLING IS MANDATORY.*
	//   - Implementations MUST implement robust error handling to manage different failure scenarios, including:
	//     - Network connectivity errors (for cloud-based OCR services). Consider implementing RETRY MECHANISMS with exponential backoff and jitter to ensure resilience to transient network issues.
	//     - API errors returned by cloud OCR services (e.g., rate limiting, authentication failures, service unavailable errors). Handle API errors GRACEFULLY and provide informative error messages.
	//     - Invalid image data or unsupported image formats. Return SPECIFIC, CUSTOM ERRORS for these cases (e.g., `domain.ErrInvalidImageFormat`) to allow for differentiated error handling in the service layer.
	//     - OCR engine-specific errors. WRAP and provide CONTEXT to these errors to aid debugging and error tracing.
	//   - Implementations SHOULD return specific, custom error types for different failure scenarios, defined in the `internal/domain/errors.go` package.
	//     Example error types to consider (and add more as needed):
	//       - `domain.ErrOCRExtractionFailed`: For general OCR extraction failures (wrap underlying errors for context).
	//       - `domain.ErrOCRServiceUnavailable`: For service unavailability or network errors with cloud OCR services (important for alerting and fallback strategies).
	//       - `domain.ErrInvalidImageFormat`: For unsupported or invalid image formats (essential for input validation and user feedback).
	//   - Implementations MUST log errors with SUFFICIENT DETAIL using a structured logger like Zap.
	//     Log entries SHOULD include:
	//       - Filename or image identifier (if available) for traceability.
	//       - Image type to aid debugging format-specific issues.
	//       - Error message from the OCR engine (if available) for root cause analysis.
	//       - Request ID and operation name for tracing and debugging in distributed systems.
	//       - Severity level (Error, Warning, Info, Debug) appropriate to the error type, following best practices for logging levels.
	//
	// Best Practices for Implementations: *ADHERE TO THESE BEST PRACTICES FOR PRODUCTION-READY OCR SERVICES.*
	//   - Robustness: Design implementations to be ROBUST and RESILIENT to handle various image qualities, orientations, and document layouts gracefully. Implement thorough INPUT VALIDATION and SANITIZATION as needed to prevent unexpected behavior or security vulnerabilities.
	//   - Performance: OPTIMIZE for LATENCY and THROUGHPUT to ensure a responsive user experience, especially under load. Consider:
	//     - Asynchronous Processing: Use goroutines and channels for NON-BLOCKING OCR processing, especially for handling multiple concurrent requests or large documents. This is HIGHLY RECOMMENDED for production deployments.
	//     - Connection Pooling: For cloud-based OCR services, use HTTP CONNECTION POOLING and KEEP-ALIVE connections to minimize latency and resource consumption. Efficient connection management is CRITICAL for performance and scalability.
	//     - Caching (Carefully Consider): EXPLORE CACHING MECHANISMS for OCR results, but only if appropriate for the application's data handling and PRIVACY REQUIREMENTS. Caching MIGHT be suitable for frequently accessed, non-PHI data, but is GENERALLY NOT RECOMMENDED for patient-sensitive medical information due to privacy and data freshness concerns. If caching is implemented, ensure CACHE INVALIDATION strategies are in place.
	//   - Security:
	//     - Secure API Key Management: If using cloud-based OCR services, ENSURE API KEYS ARE SECURELY MANAGED using environment variables or dedicated secrets management services (e.g., HashiCorp Vault, AWS Secrets Manager, Google Cloud Secret Manager). *NEVER HARDCODE API KEYS DIRECTLY IN THE CODE.* Follow the principle of least privilege when granting access to secrets.
	//     - Data Privacy:  ADHERE TO DATA PRIVACY REGULATIONS (HIPAA, GDPR, CCPA, etc.) when handling patient data with OCR services, especially cloud-based services. REVIEW the service's data processing and privacy policies CAREFULLY to ensure compliance. Implement data anonymization and de-identification techniques where appropriate.
	//   - Metrics and Monitoring:
	//     - Implement METRICS COLLECTION to MONITOR OCR PERFORMANCE in production.  This is ESSENTIAL for identifying performance bottlenecks, tracking error rates, and ensuring service reliability.  Track key metrics such as:
	//       - Latency: Request processing time for OCR operations (use histograms or percentiles to capture latency distribution).
	//       - Error Rates: Frequency of OCR extraction failures and specific error types (categorize errors by type for granular monitoring).
	//       - Accuracy (Optional and Hard to Measure): While difficult to measure directly in code, consider logging confidence scores and tracking user feedback related to OCR accuracy.  Establish baselines and track trends over time.
	//     - Use MONITORING TOOLS (e.g., Prometheus, Grafana, cloud provider's monitoring services) to VISUALIZE metrics and set up ALERTS for performance degradation or errors. Proactive monitoring and alerting are CRITICAL for maintaining service uptime and quality.
	ExtractText(ctx context.Context, imageData []byte, imageType string) (string, float64, error) // Returns text, confidence, error
}
