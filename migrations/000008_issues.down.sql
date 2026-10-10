DROP TRIGGER IF EXISTS triggerAfterChangeIssueVoters ON issue_voters;
DROP FUNCTION IF EXISTS recountIssueVotesAfterChange();
DROP TRIGGER IF EXISTS triggerAfterChangeIssueComments ON issue_comments;
DROP FUNCTION IF EXISTS recountIssueCommentsAfterChange();

DROP TABLE IF EXISTS sessions CASCADE;
DROP TABLE IF EXISTS issue_voters CASCADE;
DROP TABLE IF EXISTS issue_watchers CASCADE;
DROP TABLE IF EXISTS issue_attachments CASCADE;
DROP TABLE IF EXISTS issue_links CASCADE;
DROP TABLE IF EXISTS issue_custom_fields CASCADE;
DROP TABLE IF EXISTS issue_comments CASCADE;
DROP TABLE IF EXISTS issues CASCADE;
DROP TYPE IF EXISTS issue_link_direction CASCADE;
