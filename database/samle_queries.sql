-- =============================================================================
-- EXAMPLE QUERIES DEMONSTRATING SYSTEM CAPABILITIES
-- =============================================================================

-- Query 1: Get all children with their current teachers and groups
SELECT
    c.first_name || ' ' || c.last_name AS child_name,
    c.birthdate,
    t.first_name || ' ' || t.last_name AS current_teacher,
    c.family_language,
    c.migration_background
FROM children c
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
    COUNT(de.entry_id) AS total_documentation_entries,
    SUM(CASE WHEN de.approved = 1 THEN 1 ELSE 0 END) AS approved_entries
FROM children c
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
    t.teacher_id,
    t.first_name || ' ' || t.last_name AS teacher_name,
    cta.start_date AS assignment_start_date
FROM children c
JOIN child_teacher_assignments cta ON c.child_id = cta.child_id AND cta.end_date IS NULL
JOIN teachers t ON cta.teacher_id = t.teacher_id;

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
