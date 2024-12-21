CREATE INDEX idx_users_fullname ON users(fullname);
CREATE INDEX idx_departments_name ON departments(name);
CREATE INDEX idx_groups_name ON `groups`(name);
CREATE INDEX idx_analyses_year_semester ON analyses(year, semester);
CREATE INDEX idx_analyses_subject ON analyses(subject);