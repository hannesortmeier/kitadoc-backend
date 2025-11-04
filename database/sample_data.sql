-- =============================================================================
-- SAMPLE DATA INSERTION
-- =============================================================================

-- Insert Categories
INSERT INTO categories (category_name, description) VALUES
    ('Bewegung', 'Beobachtungen zur Bewegungsfreude, Koordination, Grundbewegungen (Robben, Klettern, Springen, Balancieren etc.) und Selbstständigkeit bei motorischen Aufgaben.'),
    ('Körper, Gesundheit, Ernährung', 'Körperwahrnehmung, Körperschema, Spannungsverhalten, Essverhalten, Gesundheitsfragen, U-Untersuchungen und Impfstatus.'),
    ('Sprache und Kommunikation', 'Sprachgebrauch, Lautbildung, Wortschatz, Erzählen, Hörverständnis, Zuhören, Grammatik und frühe Schriftsprache.'),
    ('Soziale und (inter-) kulturelle Bildung', 'Sozialverhalten in Gruppen und gegenüber Erwachsenen, Trennung, Spielverhalten, Kooperation, Konfliktlösung, Empathie und interkulturelle Anpassung.'),
    ('Musisch- ästhetische Bildung', 'Kreativität beim Gestalten, Umgang mit Farben und Materialien, Musizieren, Rhythmusgefühl und Gedächtnis für Lieder/Reime.'),
    ('Religion und Ethik', 'Interesse an religiösen Ritualen und Festen, Kenntnis biblischer Geschichten, Gerechtigkeitssinn, Solidarität und philosophische Fragen zu Leben und Tod.'),
    ('Mathematische Bildung', 'Zahlenverständnis, Mengenverständnis, Puzzeln, räumliches Vorstellungsvermögen, Vergleichen (mehr/weniger) und erste mathematische Zusammenhänge.'),
    ('Naturwissenschaftlich- technische Bildung', 'Neugier für Natur und Technik, Experimentieren mit Materialien, Beobachtung von Prozessen und Teilen von Wissen.'),
    ('Ökologische Bildung', 'Umweltbewusstsein, Kreisläufe der Natur, Trennen/Recycle von Rohstoffen und nachhaltiges Verhalten.'),
    ('Medien', 'Umgang mit Bilderbüchern und digitalen Medien, Zuhören bei Geschichten, Wiedergeben/Weitererzählen und kreativer Medieneinsatz.'),
    ('Eingewöhnung', 'Trennungs- und Bindungsfähigkeit, Erkundungsverhalten in der Kita, Nähe-Distanz-Regulation und Wohlbefinden.'),
    ('Inklusion', 'Orientierung an Teil- und Förderplan, individuelle Förderung und inklusive Unterstützung.');

-- Insert Teachers
INSERT INTO teachers (first_name, last_name, username) VALUES
    ('Maria', 'Schmidt', 'maria.schmidt'),
    ('Anna', 'Müller', 'anna.mueller'),
    ('Thomas', 'Weber', 'thomas.weber'),
    ('Sarah', 'Fischer', 'sarah.fischer'),
    ('Michael', 'Wagner', 'michael.wagner');

-- Insert Children
INSERT INTO children (first_name, last_name, birthdate, gender, family_language, migration_background, admission_date, expected_school_enrollment, address, parent1_name, parent2_name) VALUES
    ('Anna', 'Müller', '2019-03-15', 'weiblich', 'Deutsch', 0, '2023-08-01', '2025-08-01', 'Musterstraße 12, 12345 Berlin', 'Peter Johnson', 'Lisa Johnson'),
    ('Liam', 'Kowalski', '2018-11-20', 'männlich', 'Polnisch', 1, '2023-08-01', '2024-08-01', 'Hauptstraße 45, 12345 Berlin', 'Jan Kowalski', 'Anna Kowalski'),
    ('Ben', 'Springer', '2019-07-08', 'männlich', 'Spanisch', 1, '2023-09-15', '2025-08-01', 'Parkweg 23, 12345 Berlin', 'Carlos Martinez', 'Elena Martinez'),
    ('Noah', 'Brown', '2019-01-12', 'männlich', 'Englisch', 1, '2023-08-01', '2025-08-01', 'Lindenallee 67, 12345 Berlin', 'James Brown', 'Emma Brown'),
    ('Mia', 'Schneider', '2018-09-30', 'divers', 'Deutsch', 0, '2023-08-01', '2024-08-01', 'Rosenstraße 89, 12345 Berlin', 'Klaus Schneider', 'Petra Schneider'),
    ('Lucas', 'Ahmed', '2019-05-22', 'männlich', 'Arabisch', 1, '2023-10-01', '2025-08-01', 'Friedensplatz 34, 12345 Berlin', 'Omar Ahmed', 'Fatima Ahmed'),
    ('Charlotte', 'Becker', '2019-02-18', 'weiblich', 'Deutsch', 0, '2023-08-01', '2025-08-01', 'Kastanienweg 56, 12345 Berlin', 'Frank Becker', 'Sabine Becker'),
    ('Oliver', 'Popovic', '2018-12-05', 'männlich', 'Serbisch', 1, '2023-09-01', '2024-08-01', 'Eichenstraße 78, 12345 Berlin', 'Marko Popovic', 'Milica Popovic');

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
    (1, 1, 1, 'Anna zeigt große Hilfsbereitschaft gegenüber anderen Kindern. Sie hilft beim Aufräumen und tröstet weinende Kinder.', '2024-01-15', 1, 1, '2024-01-16 10:30:00'),
    (1, 1, 2, 'Anna verwendet komplexe Sätze und erzählt zusammenhängende Geschichten. Ihr Wortschatz erweitert sich täglich.', '2024-02-10', 1, 1, '2024-02-11 14:20:00'),
    (2, 1, 2, 'Liam macht große Fortschritte in der deutschen Sprache. Er kommuniziert zunehmend auf Deutsch mit anderen Kindern.', '2024-01-20', 1, 1, '2024-01-21 09:15:00'),
    (3, 2, 4, 'Ben löst Puzzles mit 50+ Teilen selbständig und zeigt dabei große Ausdauer und logisches Denken.', '2024-02-05', 1, 2, '2024-02-06 11:45:00'),
    (4, 2, 3, 'Noah zeigt ausgezeichnete Feinmotorik beim Basteln und kann bereits seinen Namen schreiben.', '2024-01-25', 0, NULL, NULL),
    (5, 3, 6, 'Mia reguliert ihre Emotionen sehr gut und kann Konflikte verbal lösen, anstatt zu weinen oder zu schreien.', '2024-02-12', 1, 3, '2024-02-13 08:30:00'),
    (6, 3, 1, 'Lucas integriert sich gut in die Gruppe und hat bereits enge Freundschaften entwickelt.', '2024-01-30', 1, 3, '2024-01-30 09:00:00'),
    (7, 4, 5, 'Charlotte zeigt große Kreativität beim Malen und Basteln. Ihre Kunstwerke sind sehr detailreich und fantasievoll.', '2024-02-08', 1, 4, '2024-02-09 13:20:00'),
    (8, 4, 3, 'Oliver turnt gerne und zeigt gute Koordination beim Klettern und Balancieren.', '2024-02-01', 0, NULL, NULL),
    (1, 1, 4, 'Anna zeigt Interesse an mathematischen Konzepten und kann bis 20 zählen.', '2024-02-20', 0, NULL, NULL);
