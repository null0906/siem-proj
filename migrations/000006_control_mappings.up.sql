CREATE TABLE control_mappings (
    id                BIGSERIAL PRIMARY KEY,
    framework         TEXT NOT NULL CHECK (framework IN ('SOC 2','ISO 27001','DPDPA')),
    control_code      TEXT NOT NULL,
    control_title     TEXT NOT NULL,
    control_summary   TEXT NOT NULL DEFAULT '',
    gap_type          TEXT NOT NULL,
    remediation_label TEXT NOT NULL,
    remediation_href  TEXT NOT NULL,
    UNIQUE (framework, control_code, gap_type)
);

CREATE INDEX idx_control_mappings_framework ON control_mappings(framework, control_code);

INSERT INTO control_mappings
    (framework, control_code, control_title, control_summary, gap_type, remediation_label, remediation_href)
VALUES
    ('SOC 2', 'CC6.1', 'Logical access safeguards', 'Access to systems is protected using strong authentication.', 'accounts_without_mfa', 'accounts lack MFA', '/identity'),
    ('SOC 2', 'CC6.1', 'Logical access safeguards', 'Access to systems is protected using strong authentication.', 'privileged_without_mfa', 'privileged accounts lack MFA', '/identity'),
    ('SOC 2', 'CC6.6', 'External access restrictions', 'Internet-facing access is limited to approved business needs.', 'public_cloud_exposure', 'public exposures remain open', '/findings?status=open'),
    ('SOC 2', 'CC6.7', 'Protection of confidential data', 'Data stored on managed devices is protected.', 'unencrypted_devices', 'devices are unencrypted', '/assets'),
    ('SOC 2', 'CC7.1', 'Vulnerability management', 'Security weaknesses are identified and remediated on time.', 'sla_breached_criticals', 'high-risk findings breached SLA', '/findings?source_tool=scanner&status=open'),
    ('SOC 2', 'CC7.1', 'Vulnerability management', 'Security weaknesses are identified and remediated on time.', 'exploitable_unpatched', 'exploitable findings remain unpatched', '/findings?source_tool=scanner&status=open'),
    ('SOC 2', 'CC7.2', 'Endpoint monitoring', 'Managed devices are monitored for security events.', 'devices_without_edr', 'devices lack active monitoring', '/assets'),
    ('ISO 27001', 'A.5.15', 'Access control', 'Access rules are established and enforced.', 'privileged_without_mfa', 'privileged accounts lack MFA', '/identity'),
    ('ISO 27001', 'A.5.18', 'Access rights review', 'Elevated and inactive access is reviewed regularly.', 'dormant_privileged', 'dormant privileged accounts need review', '/identity'),
    ('ISO 27001', 'A.8.7', 'Protection against malware', 'Endpoints use active detection and protection.', 'devices_without_edr', 'devices lack active monitoring', '/assets'),
    ('ISO 27001', 'A.8.8', 'Technical vulnerability management', 'Technical vulnerabilities are remediated based on risk.', 'open_critical_vulnerabilities', 'critical vulnerabilities remain open', '/findings?severity=critical&source_tool=scanner'),
    ('ISO 27001', 'A.8.8', 'Technical vulnerability management', 'Technical vulnerabilities are remediated based on risk.', 'sla_breached_criticals', 'high-risk findings breached SLA', '/findings?source_tool=scanner&status=open'),
    ('ISO 27001', 'A.8.24', 'Use of cryptography', 'Sensitive data is protected using appropriate cryptography.', 'unencrypted_devices', 'devices are unencrypted', '/assets'),
    ('ISO 27001', 'A.8.20', 'Network security', 'External exposure is restricted and monitored.', 'public_cloud_exposure', 'public exposures remain open', '/findings?status=open'),
    ('DPDPA', '§8(5)', 'Reasonable security safeguards', 'Personal data is protected with reasonable technical safeguards.', 'unencrypted_devices', 'devices are unencrypted', '/assets'),
    ('DPDPA', '§8(5)', 'Reasonable security safeguards', 'Personal data is protected with reasonable technical safeguards.', 'devices_without_edr', 'devices lack active monitoring', '/assets'),
    ('DPDPA', '§8(5)', 'Reasonable security safeguards', 'Personal data is protected with reasonable technical safeguards.', 'public_cloud_exposure', 'public exposures remain open', '/findings?status=open'),
    ('DPDPA', '§8(6)', 'Breach prevention and response', 'Material security weaknesses are addressed promptly.', 'open_critical_vulnerabilities', 'critical vulnerabilities remain open', '/findings?severity=critical&source_tool=scanner'),
    ('DPDPA', '§8(6)', 'Breach prevention and response', 'Material security weaknesses are addressed promptly.', 'exploitable_unpatched', 'exploitable findings remain unpatched', '/findings?source_tool=scanner&status=open'),
    ('DPDPA', '§8(7)', 'Data access governance', 'Access to personal data is limited and reviewed.', 'accounts_without_mfa', 'accounts lack MFA', '/identity'),
    ('DPDPA', '§8(7)', 'Data access governance', 'Access to personal data is limited and reviewed.', 'iam_overpermission', 'cloud access is over-permissioned', '/findings?status=open');
