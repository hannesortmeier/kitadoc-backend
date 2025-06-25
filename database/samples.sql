-- =============================================================================
-- SAMPLE DATA INSERTION
-- =============================================================================

-- Insert Groups
INSERT INTO groups (group_name) VALUES
    ('Regenbogen Gruppe'),
    ('Sonnenblumen Gruppe'),
    ('Sterne Gruppe'),
    ('Marienkäfer Gruppe');

-- Insert Categories
INSERT INTO categories (category_name, description) VALUES
    ('Soziale Entwicklung', 'Beobachtungen zur sozialen Interaktion und Entwicklung'),
    ('Sprachentwicklung', 'Beobachtungen zur sprachlichen Entwicklung'),
    ('Motorische Entwicklung', 'Beobachtungen zur körperlichen und motorischen Entwicklung'),
    ('Kognitive Entwicklung', 'Beobachtungen zur geistigen und kognitiven Entwicklung'),
    ('Kreativität', 'Beobachtungen zu kreativen Aktivitäten und Ausdrucksformen'),
    ('Emotionale Entwicklung', 'Beobachtungen zur emotionalen Reife und Regulation');

-- Insert Teachers
INSERT INTO teachers (first_name, last_name) VALUES
    ('Maria', 'Schmidt'),
    ('Anna', 'Müller'),
    ('Thomas', 'Weber'),
    ('Sarah', 'Fischer'),
    ('Michael', 'Wagner');

-- Insert Children
INSERT INTO children (first_name, last_name, birthdate, group_id, family_language, migration_background, admission_date, expected_school_enrollment, address, parent1_name, parent2_name) VALUES
    ('Emma', 'Johnson', '2019-03-15', 1, 'Deutsch', 0, '2023-08-01', '2025-08-01', 'Musterstraße 12, 12345 Berlin', 'Peter Johnson', 'Lisa Johnson'),
    ('Liam', 'Kowalski', '2018-11-20', 1, 'Polnisch', 1, '2023-08-01', '2024-08-01', 'Hauptstraße 45, 12345 Berlin', 'Jan Kowalski', 'Anna Kowalski'),
    ('Sophie', 'Martinez', '2019-07-08', 2, 'Spanisch', 1, '2023-09-15', '2025-08-01', 'Parkweg 23, 12345 Berlin', 'Carlos Martinez', 'Elena Martinez'),
    ('Noah', 'Brown', '2019-01-12', 2, 'Englisch', 1, '2023-08-01', '2025-08-01', 'Lindenallee 67, 12345 Berlin', 'James Brown', 'Emma Brown'),
    ('Mia', 'Schneider', '2018-09-30', 3, 'Deutsch', 0, '2023-08-01', '2024-08-01', 'Rosenstraße 89, 12345 Berlin', 'Klaus Schneider', 'Petra Schneider'),
    ('Lucas', 'Ahmed', '2019-05-22', 3, 'Arabisch', 1, '2023-10-01', '2025-08-01', 'Friedensplatz 34, 12345 Berlin', 'Omar Ahmed', 'Fatima Ahmed'),
    ('Charlotte', 'Becker', '2019-02-18', 4, 'Deutsch', 0, '2023-08-01', '2025-08-01', 'Kastanienweg 56, 12345 Berlin', 'Frank Becker', 'Sabine Becker'),
    ('Oliver', 'Popovic', '2018-12-05', 4, 'Serbisch', 1, '2023-09-01', '2024-08-01', 'Eichenstraße 78, 12345 Berlin', 'Marko Popovic', 'Milica Popovic');

-- Insert Child-Teacher Assignments
INSERT INTO child_teacher_assignments (child_id, teacher_id, start_date, end_date) VALUES
    (1, 1, '2023-08-01', NULL),  -- Emma with Maria Schmidt (current)
    (2, 1, '2023-08-01', NULL),  -- Liam with Maria Schmidt (current)
    (3, 2, '2023-09-15', NULL),  -- Sophie with Anna Müller (current)
    (4, 2, '2023-08-01', NULL),  -- Noah with Anna Müller (current)
    (5, 3, '2023-08-01', NULL),  -- Mia with Thomas Weber (current)
    (6, 3, '2023-10-01', NULL),  -- Lucas with Thomas Weber (current)
    (7, 4, '2023-08-01', NULL),  -- Charlotte with Sarah Fischer (current)
    (8, 4, '2023-09-01', NULL),  -- Oliver with Sarah Fischer (current)
    (1, 5, '2023-08-01', '2023-12-15'),  -- Emma had Michael Wagner initially
    (3, 1, '2023-09-15', '2023-11-30');  -- Sophie had Maria Schmidt briefly

-- Insert Documentation Entries
INSERT INTO documentation_entries (child_id, documenting_teacher_id, category_id, observation_description, observation_date, approved, approved_by_teacher_id, approved_at) VALUES
    (1, 1, 1, 'Emma zeigt große Hilfsbereitschaft gegenüber anderen Kindern. Sie hilft beim Aufräumen und tröstet weinende Kinder.', '2024-01-15', 1, 5, '2024-01-16 10:30:00'),
    (1, 1, 2, 'Emma verwendet komplexe Sätze und erzählt zusammenhängende Geschichten. Ihr Wortschatz erweitert sich täglich.', '2024-02-10', 1, 5, '2024-02-11 14:20:00'),
    (2, 1, 2, 'Liam macht große Fortschritte in der deutschen Sprache. Er kommuniziert zunehmend auf Deutsch mit anderen Kindern.', '2024-01-20', 1, 5, '2024-01-21 09:15:00'),
    (3, 2, 4, 'Sophie löst Puzzles mit 50+ Teilen selbständig und zeigt dabei große Ausdauer und logisches Denken.', '2024-02-05', 1, 3, '2024-02-06 11:45:00'),
    (4, 2, 3, 'Noah zeigt ausgezeichnete Feinmotorik beim Basteln und kann bereits seinen Namen schreiben.', '2024-01-25', 0, NULL, NULL),
    (5, 3, 6, 'Mia reguliert ihre Emotionen sehr gut und kann Konflikte verbal lösen, anstatt zu weinen oder zu schreien.', '2024-02-12', 1, 1, '2024-02-13 08:30:00'),
    (6, 3, 1, 'Lucas integriert sich gut in die Gruppe und hat bereits enge Freundschaften entwickelt.', '2024-01-30', 0, NULL, NULL),
    (7, 4, 5, 'Charlotte zeigt große Kreativität beim Malen und Basteln. Ihre Kunstwerke sind sehr detailreich und fantasievoll.', '2024-02-08', 1, 2, '2024-02-09 13:20:00'),
    (8, 4, 3, 'Oliver turnt gerne und zeigt gute Koordination beim Klettern und Balancieren.', '2024-02-01', 0, NULL, NULL),
    (1, 1, 4, 'Emma zeigt Interesse an mathematischen Konzepten und kann bis 20 zählen.', '2024-02-20', 0, NULL, NULL);

-- =============================================================================
-- EXAMPLE QUERIES DEMONSTRATING SYSTEM CAPABILITIES
-- =============================================================================

-- Query 1: Get all children with their current teachers and groups
SELECT
    c.first_name || ' ' || c.last_name AS child_name,
    c.birthdate,
    g.group_name,
    t.first_name || ' ' || t.last_name AS current_teacher,
    c.family_language,
    c.migration_background
FROM children c
LEFT JOIN groups g ON c.group_id = g.group_id
LEFT JOIN child_teacher_assignments cta ON c.child_id = cta.child_id AND cta.end_date IS NULL
LEFT JOIN teachers t ON cta.teacher_id = t.teacher_id
ORDER BY c.last_name, c.first_name;

-- Query 2: Get all documentation entries for a specific child with approval status
SELECT
    de.observation_date,
    cat.category_name,
    de.observation_description,
    t1.first_name || ' ' || t1.last_name AS documenting_teacher,
    CASE
        WHEN de.approved = 1 THEN 'Genehmigt von ' || t2.first_name || ' ' || t2.last_name
        ELSE 'Noch nicht genehmigt'
    END AS approval_status,
    de.created_at
FROM documentation_entries de
JOIN teachers t1 ON de.documenting_teacher_id = t1.teacher_id
LEFT JOIN teachers t2 ON de.approved_by_teacher_id = t2.teacher_id
LEFT JOIN categories cat ON de.category_id = cat.category_id
WHERE de.child_id = 1  -- Emma Johnson
ORDER BY de.observation_date DESC;

-- Query 3: Get teacher workload - count of children currently assigned to each teacher
SELECT
    t.first_name || ' ' || t.last_name AS teacher_name,
    COUNT(cta.child_id) AS current_children_count,
    GROUP_CONCAT(c.first_name || ' ' || c.last_name, ', ') AS children_names
FROM teachers t
LEFT JOIN child_teacher_assignments cta ON t.teacher_id = cta.teacher_id AND cta.end_date IS NULL
LEFT JOIN children c ON cta.child_id = c.child_id
GROUP BY t.teacher_id, t.first_name, t.last_name
ORDER BY current_children_count DESC;

-- Query 4: Get children ready for school enrollment (born in 2018)
SELECT
    c.first_name || ' ' || c.last_name AS child_name,
    c.birthdate,
    c.expected_school_enrollment,
    g.group_name,
    COUNT(de.entry_id) AS total_documentation_entries,
    SUM(CASE WHEN de.approved = 1 THEN 1 ELSE 0 END) AS approved_entries
FROM children c
LEFT JOIN groups g ON c.group_id = g.group_id
LEFT JOIN documentation_entries de ON c.child_id = de.child_id
WHERE strftime('%Y', c.birthdate) = '2018'
GROUP BY c.child_id
ORDER BY c.expected_school_enrollment;

-- Query 5: Get documentation statistics by category and month
SELECT
    strftime('%Y-%m', de.observation_date) AS observation_month,
    cat.category_name,
    COUNT(de.entry_id) AS total_entries,
    SUM(CASE WHEN de.approved = 1 THEN 1 ELSE 0 END) AS approved_entries,
    ROUND(AVG(CASE WHEN de.approved = 1 THEN 1.0 ELSE 0.0 END) * 100, 2) AS approval_percentage
FROM documentation_entries de
LEFT JOIN categories cat ON de.category_id = cat.category_id
WHERE de.observation_date >= DATE('now', '-6 months')
GROUP BY strftime('%Y-%m', de.observation_date), cat.category_id
ORDER BY observation_month DESC, cat.category_name;

-- =============================================================================
-- ADDITIONAL UTILITY VIEWS
-- =============================================================================

-- View for current child-teacher assignments
CREATE VIEW current_assignments AS
SELECT
    c.child_id,
    c.first_name || ' ' || c.last_name AS child_name,
    c.birthdate,
    g.group_name,
    t.teacher_id,
    t.first_name || ' ' || t.last_name AS teacher_name,
    cta.start_date AS assignment_start_date
FROM children c
JOIN child_teacher_assignments cta ON c.child_id = cta.child_id AND cta.end_date IS NULL
JOIN teachers t ON cta.teacher_id = t.teacher_id
LEFT JOIN groups g ON c.group_id = g.group_id;

-- View for documentation summary per child
CREATE VIEW child_documentation_summary AS
SELECT
    c.child_id,
    c.first_name || ' ' || c.last_name AS child_name,
    COUNT(de.entry_id) AS total_entries,
    SUM(CASE WHEN de.approved = 1 THEN 1 ELSE 0 END) AS approved_entries,
    COUNT(DISTINCT de.category_id) AS categories_documented,
    MAX(de.observation_date) AS last_documentation_date
FROM children c
LEFT JOIN documentation_entries de ON c.child_id = de.child_id
GROUP BY c.child_id;
