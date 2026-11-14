-- +goose Up
-- Seed default Salesforce admin user
INSERT INTO salesforce.salesforce_user (id, organization_id, username, email, password_hash, name, first_name, last_name, nickname, is_active)
VALUES ('005000000000001AAA', '00D000000000001EAA', 'admin@example.com', 'admin@example.com', 'password123', 'System Administrator', 'System', 'Administrator', 'admin', TRUE)
ON CONFLICT (username) DO NOTHING;

-- Seed default OAuth client (matching Bruno collection default clientId + generic test client)
INSERT INTO salesforce.oauth_client (id, client_id, client_secret, name, redirect_uri, scopes, is_active)
VALUES ('0H4000000000001AAA', '3MVG98_Psg5cppyYCmk1gZNC25o00SXpgpodlS29IZ6pXiHkt3xuPa5qIjBTtEgdsiMuIWVN_8F0jnwEtbDh4', 'secret', 'Salesforce Platform APIs Client', 'http://localhost:8080/oauth2/callback', 'api refresh_token openid', TRUE),
       ('0H4000000000002AAA', 'test_client_id', 'test_client_secret', 'Test Connected App', 'http://localhost:8080/callback', 'api refresh_token openid', TRUE)
ON CONFLICT (client_id) DO NOTHING;

-- Seed Organization Limits
INSERT INTO salesforce.org_limit (limit_name, max_limit, remaining_limit)
VALUES ('DailyApiRequests', 15000, 14995),
       ('DailyBulkApiRequests', 10000, 10000),
       ('DailyBulkV2QueryJobs', 10000, 10000),
       ('DailyBulkV2QueryFileStorageMB', 976562, 976562),
       ('DataStorageMB', 1024, 1000),
       ('FileStorageMB', 1024, 1020),
       ('SingleEmail', 5000, 5000),
       ('MassEmail', 5000, 5000),
       ('DailyStreamingApiEvents', 10000, 10000),
       ('DailyStandardVolumePlatformEvents', 25000, 25000)
ON CONFLICT (limit_name) DO NOTHING;

-- Seed SObject Definitions (Standard, Custom, and Tooling Objects)
INSERT INTO salesforce.sobject_definition (name, label, label_plural, key_prefix, custom, is_tooling, table_name)
VALUES ('Account', 'Account', 'Accounts', '001', FALSE, FALSE, 'account'),
       ('Contact', 'Contact', 'Contacts', '003', FALSE, FALSE, 'contact'),
       ('Lead', 'Lead', 'Leads', '00Q', FALSE, FALSE, 'lead'),
       ('Case', 'Case', 'Cases', '500', FALSE, FALSE, 'case'),
       ('Custom_Object__c', 'Custom Object', 'Custom Objects', 'a00', TRUE, FALSE, 'custom_object__c'),
       ('ContentDocument', 'Content Document', 'Content Documents', '069', FALSE, FALSE, 'content_document'),
       ('ContentVersion', 'Content Version', 'Content Versions', '068', FALSE, FALSE, 'content_version'),
       ('ContentDocumentLink', 'Content Document Link', 'Content Document Links', '06A', FALSE, FALSE, 'content_document_link'),
       ('ApexClass', 'Apex Class', 'Apex Classes', '01p', FALSE, TRUE, 'apex_class'),
       ('ApexTrigger', 'Apex Trigger', 'Apex Triggers', '01q', FALSE, TRUE, 'apex_trigger'),
       ('CustomField', 'Custom Field', 'Custom Fields', '00N', FALSE, TRUE, 'custom_field'),
       ('ValidationRule', 'Validation Rule', 'Validation Rules', '03d', FALSE, TRUE, 'validation_rule')
ON CONFLICT (name) DO NOTHING;

-- Seed SObject Field Definitions
INSERT INTO salesforce.sobject_field_definition (sobject_name, name, column_name, label, field_type, soap_type, length, nillable, createable, updateable, custom, external_id, name_field)
VALUES ('Account', 'Id', 'id', 'Account ID', 'id', 'tns:ID', 18, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE),
       ('Account', 'External_ID__c', 'external_id__c', 'External ID', 'string', 'xsd:string', 255, TRUE, TRUE, TRUE, TRUE, TRUE, FALSE),
       ('Account', 'Name', 'name', 'Account Name', 'string', 'xsd:string', 255, FALSE, TRUE, TRUE, FALSE, FALSE, TRUE),
       ('Account', 'Type', 'type', 'Account Type', 'picklist', 'xsd:string', 40, TRUE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('Account', 'Industry', 'industry', 'Industry', 'picklist', 'xsd:string', 40, TRUE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('Account', 'Phone', 'phone', 'Account Phone', 'phone', 'xsd:string', 40, TRUE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('Account', 'Website', 'website', 'Website', 'url', 'xsd:string', 255, TRUE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('Account', 'BillingCity', 'billing_city', 'Billing City', 'string', 'xsd:string', 40, TRUE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('Account', 'BillingState', 'billing_state', 'Billing State/Province', 'string', 'xsd:string', 80, TRUE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('Account', 'BillingCountry', 'billing_country', 'Billing Country', 'string', 'xsd:string', 80, TRUE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('Account', 'Description', 'description', 'Account Description', 'textarea', 'xsd:string', 32000, TRUE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('Account', 'CreatedDate', 'created_date', 'Created Date', 'datetime', 'xsd:dateTime', 0, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE),
       ('Contact', 'Id', 'id', 'Contact ID', 'id', 'tns:ID', 18, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE),
       ('Contact', 'External_ID__c', 'external_id__c', 'External ID', 'string', 'xsd:string', 255, TRUE, TRUE, TRUE, TRUE, TRUE, FALSE),
       ('Contact', 'AccountId', 'account_id', 'Account ID', 'reference', 'tns:ID', 18, TRUE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('Contact', 'FirstName', 'first_name', 'First Name', 'string', 'xsd:string', 40, TRUE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('Contact', 'LastName', 'last_name', 'Last Name', 'string', 'xsd:string', 80, FALSE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('Contact', 'Name', 'name', 'Full Name', 'string', 'xsd:string', 121, FALSE, FALSE, FALSE, FALSE, FALSE, TRUE),
       ('Contact', 'Email', 'email', 'Email', 'email', 'xsd:string', 80, TRUE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('Contact', 'Phone', 'phone', 'Business Phone', 'phone', 'xsd:string', 40, TRUE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('Contact', 'Title', 'title', 'Title', 'string', 'xsd:string', 128, TRUE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('Contact', 'Department', 'department', 'Department', 'string', 'xsd:string', 80, TRUE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('Contact', 'CreatedDate', 'created_date', 'Created Date', 'datetime', 'xsd:dateTime', 0, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE),
       ('Lead', 'Id', 'id', 'Lead ID', 'id', 'tns:ID', 18, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE),
       ('Lead', 'External_ID__c', 'external_id__c', 'External ID', 'string', 'xsd:string', 255, TRUE, TRUE, TRUE, TRUE, TRUE, FALSE),
       ('Lead', 'FirstName', 'first_name', 'First Name', 'string', 'xsd:string', 40, TRUE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('Lead', 'LastName', 'last_name', 'Last Name', 'string', 'xsd:string', 80, FALSE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('Lead', 'Name', 'name', 'Full Name', 'string', 'xsd:string', 121, FALSE, FALSE, FALSE, FALSE, FALSE, TRUE),
       ('Lead', 'Company', 'company', 'Company', 'string', 'xsd:string', 255, FALSE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('Lead', 'Status', 'status', 'Status', 'picklist', 'xsd:string', 40, FALSE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('Lead', 'Email', 'email', 'Email', 'email', 'xsd:string', 80, TRUE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('Lead', 'Phone', 'phone', 'Phone', 'phone', 'xsd:string', 40, TRUE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('Lead', 'CreatedDate', 'created_date', 'Created Date', 'datetime', 'xsd:dateTime', 0, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE),
       ('Case', 'Id', 'id', 'Case ID', 'id', 'tns:ID', 18, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE),
       ('Case', 'External_ID__c', 'external_id__c', 'External ID', 'string', 'xsd:string', 255, TRUE, TRUE, TRUE, TRUE, TRUE, FALSE),
       ('Case', 'CaseNumber', 'case_number', 'Case Number', 'string', 'xsd:string', 30, FALSE, FALSE, FALSE, FALSE, FALSE, TRUE),
       ('Case', 'Subject', 'subject', 'Subject', 'string', 'xsd:string', 255, TRUE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('Case', 'Status', 'status', 'Status', 'picklist', 'xsd:string', 40, TRUE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('Case', 'Priority', 'priority', 'Priority', 'picklist', 'xsd:string', 40, FALSE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('Case', 'Origin', 'origin', 'Case Origin', 'picklist', 'xsd:string', 40, TRUE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('Case', 'Description', 'description', 'Description', 'textarea', 'xsd:string', 32000, TRUE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('Case', 'AccountId', 'account_id', 'Account ID', 'reference', 'tns:ID', 18, TRUE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('Case', 'CreatedDate', 'created_date', 'Created Date', 'datetime', 'xsd:dateTime', 0, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE),
       ('Custom_Object__c', 'Id', 'id', 'Record ID', 'id', 'tns:ID', 18, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE),
       ('Custom_Object__c', 'External_ID__c', 'external_id__c', 'External ID', 'string', 'xsd:string', 255, TRUE, TRUE, TRUE, TRUE, TRUE, FALSE),
       ('Custom_Object__c', 'Name', 'name', 'Custom Object Name', 'string', 'xsd:string', 80, FALSE, TRUE, TRUE, FALSE, FALSE, TRUE),
       ('Custom_Object__c', 'Field_To_Update__c', 'field_to_update__c', 'Field To Update', 'string', 'xsd:string', 255, TRUE, TRUE, TRUE, TRUE, FALSE, FALSE),
       ('Custom_Object__c', 'Description__c', 'description__c', 'Description', 'textarea', 'xsd:string', 32000, TRUE, TRUE, TRUE, TRUE, FALSE, FALSE),
       ('Custom_Object__c', 'Status__c', 'status__c', 'Status', 'picklist', 'xsd:string', 40, TRUE, TRUE, TRUE, TRUE, FALSE, FALSE),
       ('Custom_Object__c', 'CreatedDate', 'created_date', 'Created Date', 'datetime', 'xsd:dateTime', 0, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE),
       ('ApexClass', 'Id', 'id', 'Apex Class ID', 'id', 'tns:ID', 18, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE),
       ('ApexClass', 'Name', 'name', 'Name', 'string', 'xsd:string', 255, FALSE, TRUE, TRUE, FALSE, FALSE, TRUE),
       ('ApexClass', 'Body', 'body', 'Body', 'textarea', 'xsd:string', 1000000, FALSE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('ApexClass', 'ApiVersion', 'api_version', 'Api Version', 'double', 'xsd:double', 0, FALSE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('ApexClass', 'Status', 'status', 'Status', 'picklist', 'xsd:string', 40, FALSE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('ApexTrigger', 'Id', 'id', 'Trigger ID', 'id', 'tns:ID', 18, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE),
       ('ApexTrigger', 'Name', 'name', 'Name', 'string', 'xsd:string', 255, FALSE, TRUE, TRUE, FALSE, FALSE, TRUE),
       ('ApexTrigger', 'TableEnumOrId', 'table_enum_or_id', 'sObject Type', 'string', 'xsd:string', 255, FALSE, TRUE, FALSE, FALSE, FALSE, FALSE),
       ('ApexTrigger', 'Body', 'body', 'Body', 'textarea', 'xsd:string', 1000000, FALSE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('CustomField', 'Id', 'id', 'Custom Field Definition ID', 'id', 'tns:ID', 18, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE),
       ('CustomField', 'DeveloperName', 'developer_name', 'Field Name', 'string', 'xsd:string', 80, FALSE, TRUE, TRUE, FALSE, FALSE, TRUE),
       ('CustomField', 'TableEnumOrId', 'table_enum_or_id', 'Table Enum Or Id', 'string', 'xsd:string', 255, FALSE, TRUE, FALSE, FALSE, FALSE, FALSE),
       ('CustomField', 'FullName', 'full_name', 'Full Name', 'string', 'xsd:string', 255, FALSE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('ValidationRule', 'Id', 'id', 'Validation Rule ID', 'id', 'tns:ID', 18, FALSE, FALSE, FALSE, FALSE, FALSE, FALSE),
       ('ValidationRule', 'ValidationName', 'validation_name', 'Rule Name', 'string', 'xsd:string', 80, FALSE, TRUE, TRUE, FALSE, FALSE, TRUE),
       ('ValidationRule', 'Active', 'active', 'Active', 'boolean', 'xsd:boolean', 0, FALSE, TRUE, TRUE, FALSE, FALSE, FALSE),
       ('ValidationRule', 'ErrorMessage', 'error_message', 'Error Message', 'string', 'xsd:string', 255, FALSE, TRUE, TRUE, FALSE, FALSE, FALSE)
ON CONFLICT (sobject_name, name) DO NOTHING;

-- Seed sample Tooling records
INSERT INTO salesforce.apex_class (id, name, body, api_version, status, is_valid, length_without_comments)
VALUES ('01p000000000001AAA', 'AccountService', 'public class AccountService { public static void process() {} }', 61.0, 'Active', TRUE, 63),
       ('01p000000000002AAA', 'AccountServiceTest', '@IsTest private class AccountServiceTest { @IsTest static void testProcess() { AccountService.process(); } }', 61.0, 'Active', TRUE, 108)
ON CONFLICT (name) DO NOTHING;

INSERT INTO salesforce.apex_trigger (id, name, table_enum_or_id, body, api_version, status, is_valid, usage_before_insert)
VALUES ('01q000000000001AAA', 'AccountTrigger', 'Account', 'trigger AccountTrigger on Account (before insert) {}', 61.0, 'Active', TRUE, TRUE)
ON CONFLICT (name) DO NOTHING;

INSERT INTO salesforce.custom_field (id, developer_name, table_enum_or_id, full_name, label, data_type, length)
VALUES ('00N000000000001AAA', 'Field_To_Update', 'Custom_Object__c', 'Custom_Object__c.Field_To_Update__c', 'Field To Update', 'Text', 255)
ON CONFLICT (full_name) DO NOTHING;

INSERT INTO salesforce.validation_rule (id, validation_name, entity_definition_id, active, description, error_condition_formula, error_message)
VALUES ('03d000000000001AAA', 'Require_Account_Phone', 'Account', TRUE, 'Requires phone when type is Customer', 'ISBLANK(Phone)', 'Phone is required')
ON CONFLICT (id) DO NOTHING;

-- Seed sample Apex REST routes
INSERT INTO salesforce.apex_rest_route (id, url_mapping, http_method, class_name, response_status, response_body, is_active)
VALUES ('01p000000000010AAA', 'v1/integrations/status', 'GET', 'IntegrationStatusResource', 200, '{"status":"UP","service":"salesforce-emulator","version":"v1"}', TRUE),
       ('01p000000000011AAA', 'v1/integrations/status', 'POST', 'IntegrationStatusResource', 200, '{"status":"UPDATED","service":"salesforce-emulator","version":"v1"}', TRUE)
ON CONFLICT (url_mapping, http_method) DO NOTHING;

-- Seed sample Account, Contact, Lead records for initial queries and SOSL search
INSERT INTO salesforce.account (id, external_id__c, name, type, industry, phone, website, billing_city, billing_state, billing_country, description)
VALUES ('001000000000001AAA', 'ACC-EXT-001', 'Acme Corporation', 'Customer - Direct', 'Technology', '555-0100', 'https://acme.example.com', 'San Francisco', 'CA', 'USA', 'Default seeded Acme account')
ON CONFLICT (id) DO NOTHING;

INSERT INTO salesforce.contact (id, external_id__c, account_id, first_name, last_name, name, email, phone, title, department)
VALUES ('003000000000001AAA', 'CON-EXT-001', '001000000000001AAA', 'Jane', 'Doe', 'Jane Doe', 'jane.doe@acme.example.com', '555-0101', 'VP of Engineering', 'Engineering')
ON CONFLICT (id) DO NOTHING;

INSERT INTO salesforce.lead (id, external_id__c, first_name, last_name, name, company, status, email, phone)
VALUES ('00Q000000000001AAA', 'LEAD-EXT-001', 'John', 'Prospect', 'John Prospect', 'Acme Subsidiary', 'Open - Not Contacted', 'john@acme-sub.example.com', '555-0102')
ON CONFLICT (id) DO NOTHING;

-- +goose Down
DELETE FROM salesforce.lead WHERE id = '00Q000000000001AAA';
DELETE FROM salesforce.contact WHERE id = '003000000000001AAA';
DELETE FROM salesforce.account WHERE id = '001000000000001AAA';
DELETE FROM salesforce.apex_rest_route WHERE id IN ('01p000000000010AAA', '01p000000000011AAA');
DELETE FROM salesforce.validation_rule WHERE id = '03d000000000001AAA';
DELETE FROM salesforce.custom_field WHERE id = '00N000000000001AAA';
DELETE FROM salesforce.apex_trigger WHERE id = '01q000000000001AAA';
DELETE FROM salesforce.apex_class WHERE id IN ('01p000000000001AAA', '01p000000000002AAA');
DELETE FROM salesforce.sobject_definition;
DELETE FROM salesforce.org_limit;
DELETE FROM salesforce.oauth_client WHERE id IN ('0H4000000000001AAA', '0H4000000000002AAA');
DELETE FROM salesforce.salesforce_user WHERE id = '005000000000001AAA';
