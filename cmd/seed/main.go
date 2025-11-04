package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"kitadoc-backend/data"
	"kitadoc-backend/migrations"
	"kitadoc-backend/models"
)

func main() {
	dsn := flag.String("dsn", "file:test.db?_foreign_keys=on", "SQLite DSN")
	key := flag.String("key", "0123456789abcdef0123456789abcdef", "32-byte hex encryption key (raw string)")
	flag.Parse()

	db, err := sql.Open("sqlite3", *dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close() // nolint:errcheck

	// Run migrations
	if err := data.MigrateDB(db, migrations.Files); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	// Create DAL
	dal := data.NewDAL(db, []byte(*key))

	// Seed categories
	categories := []models.Category{
		{Name: "Bewegung", Description: models.StringPtr("Beobachtungen zur Bewegungsfreude, Koordination, Grundbewegungen (Robben, Klettern, Springen, Balancieren etc.) und Selbstständigkeit bei motorischen Aufgaben.")},
		{Name: "Körper, Gesundheit, Ernährung", Description: models.StringPtr("Körperwahrnehmung, Körperschema, Spannungsverhalten, Essverhalten, Gesundheitsfragen, U-Untersuchungen und Impfstatus.")},
		{Name: "Sprache und Kommunikation", Description: models.StringPtr("Sprachgebrauch, Lautbildung, Wortschatz, Erzählen, Hörverständnis, Zuhören, Grammatik und frühe Schriftsprache.")},
		{Name: "Soziale und (inter-) kulturelle Bildung", Description: models.StringPtr("Sozialverhalten in Gruppen und gegenüber Erwachsenen, Trennung, Spielverhalten, Kooperation, Konfliktlösung, Empathie und interkulturelle Anpassung.")},
		{Name: "Musisch- ästhetische Bildung", Description: models.StringPtr("Kreativität beim Gestalten, Umgang mit Farben und Materialien, Musizieren, Rhythmusgefühl und Gedächtnis für Lieder/Reime.")},
		{Name: "Religion und Ethik", Description: models.StringPtr("Interesse an religiösen Ritualen und Festen, Kenntnis biblischer Geschichten, Gerechtigkeitssinn, Solidarität und philosophische Fragen zu Leben und Tod.")},
		{Name: "Mathematische Bildung", Description: models.StringPtr("Zahlenverständnis, Mengenverständnis, Puzzeln, räumliches Vorstellungsvermögen, Vergleichen (mehr/weniger) und erste mathematische Zusammenhänge.")},
		{Name: "Naturwissenschaftlich- technische Bildung", Description: models.StringPtr("Neugier für Natur und Technik, Experimentieren mit Materialien, Beobachtung von Prozessen und Teilen von Wissen.")},
		{Name: "Ökologische Bildung", Description: models.StringPtr("Umweltbewusstsein, Kreisläufe der Natur, Trennen/Recycle von Rohstoffen und nachhaltiges Verhalten.")},
		{Name: "Medien", Description: models.StringPtr("Umgang mit Bilderbüchern und digitalen Medien, Zuhören bei Geschichten, Wiedergeben/Weitererzählen und kreativer Medieneinsatz.")},
		{Name: "Eingewöhnung", Description: models.StringPtr("Trennungs- und Bindungsfähigkeit, Erkundungsverhalten in der Kita, Nähe-Distanz-Regulation und Wohlbefinden.")},
		{Name: "Inklusion", Description: models.StringPtr("Orientierung an Teil- und Förderplan, individuelle Förderung und inklusive Unterstützung.")},
	}

	for _, c := range categories {
		if _, err := dal.Categories.Create(&c); err != nil {
			log.Fatalf("failed to create category %s: %v", c.Name, err)
		}
	}

	// Seed teachers
	teachers := []models.Teacher{
		{FirstName: "Maria", LastName: "Schmidt", Username: "maria.schmidt"},
		{FirstName: "Anna", LastName: "Müller", Username: "anna.mueller"},
		{FirstName: "Thomas", LastName: "Weber", Username: "thomas.weber"},
		{FirstName: "Sarah", LastName: "Fischer", Username: "sarah.fischer"},
		{FirstName: "Michael", LastName: "Wagner", Username: "michael.wagner"},
	}

	for i := range teachers {
		if _, err := dal.Teachers.Create(&teachers[i]); err != nil {
			log.Fatalf("failed to create teacher %s: %v", teachers[i].Username, err)
		}
	}

	// Helper to parse date strings in sample_data.sql (YYYY-MM-DD)
	parseDate := func(s string) time.Time {
		t, _ := time.Parse("2006-01-02", s)
		return t
	}

	// Seed children
	children := []models.Child{
		{FirstName: "Anna", LastName: "Müller", Birthdate: parseDate("2019-03-15"), Gender: "weiblich", FamilyLanguage: "Deutsch", MigrationBackground: false, AdmissionDate: parseDate("2023-08-01"), ExpectedSchoolEnrollment: parseDate("2025-08-01"), Address: "Musterstraße 12, 12345 Berlin", Parent1Name: "Peter Johnson", Parent2Name: "Lisa Johnson"},
		{FirstName: "Liam", LastName: "Kowalski", Birthdate: parseDate("2018-11-20"), Gender: "männlich", FamilyLanguage: "Polnisch", MigrationBackground: true, AdmissionDate: parseDate("2023-08-01"), ExpectedSchoolEnrollment: parseDate("2024-08-01"), Address: "Hauptstraße 45, 12345 Berlin", Parent1Name: "Jan Kowalski", Parent2Name: "Anna Kowalski"},
		{FirstName: "Ben", LastName: "Springer", Birthdate: parseDate("2019-07-08"), Gender: "männlich", FamilyLanguage: "Spanisch", MigrationBackground: true, AdmissionDate: parseDate("2023-09-15"), ExpectedSchoolEnrollment: parseDate("2025-08-01"), Address: "Parkweg 23, 12345 Berlin", Parent1Name: "Carlos Martinez", Parent2Name: "Elena Martinez"},
		{FirstName: "Noah", LastName: "Brown", Birthdate: parseDate("2019-01-12"), Gender: "männlich", FamilyLanguage: "Englisch", MigrationBackground: true, AdmissionDate: parseDate("2023-08-01"), ExpectedSchoolEnrollment: parseDate("2025-08-01"), Address: "Lindenallee 67, 12345 Berlin", Parent1Name: "James Brown", Parent2Name: "Emma Brown"},
		{FirstName: "Mia", LastName: "Schneider", Birthdate: parseDate("2018-09-30"), Gender: "divers", FamilyLanguage: "Deutsch", MigrationBackground: false, AdmissionDate: parseDate("2023-08-01"), ExpectedSchoolEnrollment: parseDate("2024-08-01"), Address: "Rosenstraße 89, 12345 Berlin", Parent1Name: "Klaus Schneider", Parent2Name: "Petra Schneider"},
		{FirstName: "Lucas", LastName: "Ahmed", Birthdate: parseDate("2019-05-22"), Gender: "männlich", FamilyLanguage: "Arabisch", MigrationBackground: true, AdmissionDate: parseDate("2023-10-01"), ExpectedSchoolEnrollment: parseDate("2025-08-01"), Address: "Friedensplatz 34, 12345 Berlin", Parent1Name: "Omar Ahmed", Parent2Name: "Fatima Ahmed"},
		{FirstName: "Charlotte", LastName: "Becker", Birthdate: parseDate("2019-02-18"), Gender: "weiblich", FamilyLanguage: "Deutsch", MigrationBackground: false, AdmissionDate: parseDate("2023-08-01"), ExpectedSchoolEnrollment: parseDate("2025-08-01"), Address: "Kastanienweg 56, 12345 Berlin", Parent1Name: "Frank Becker", Parent2Name: "Sabine Becker"},
		{FirstName: "Oliver", LastName: "Popovic", Birthdate: parseDate("2018-12-05"), Gender: "männlich", FamilyLanguage: "Serbisch", MigrationBackground: true, AdmissionDate: parseDate("2023-09-01"), ExpectedSchoolEnrollment: parseDate("2024-08-01"), Address: "Eichenstraße 78, 12345 Berlin", Parent1Name: "Marko Popovic", Parent2Name: "Milica Popovic"},
	}

	for i := range children {
		if _, err := dal.Children.Create(&children[i]); err != nil {
			log.Fatalf("failed to create child %s %s: %v", children[i].FirstName, children[i].LastName, err)
		}
	}

	// Seed assignments. We need teacher and child IDs — for simplicity we'll assume insertion order and AUTOINCREMENT starting from 1
	assignments := []models.Assignment{
		{ChildID: 1, TeacherID: 1, StartDate: parseDate("2023-08-01")},
		{ChildID: 2, TeacherID: 1, StartDate: parseDate("2023-08-01")},
		{ChildID: 3, TeacherID: 2, StartDate: parseDate("2023-09-15")},
		{ChildID: 4, TeacherID: 2, StartDate: parseDate("2023-08-01")},
		{ChildID: 5, TeacherID: 3, StartDate: parseDate("2023-08-01")},
		{ChildID: 6, TeacherID: 3, StartDate: parseDate("2023-10-01")},
		{ChildID: 7, TeacherID: 4, StartDate: parseDate("2023-08-01")},
		{ChildID: 8, TeacherID: 4, StartDate: parseDate("2023-09-01")},
	}

	// Add two historical assignments with end dates
	end1 := parseDate("2023-12-15")
	end2 := parseDate("2023-11-30")
	assignments = append(assignments, models.Assignment{ChildID: 1, TeacherID: 5, StartDate: parseDate("2023-08-01"), EndDate: &end1})
	assignments = append(assignments, models.Assignment{ChildID: 3, TeacherID: 1, StartDate: parseDate("2023-09-15"), EndDate: &end2})

	for i := range assignments {
		if _, err := dal.Assignments.Create(&assignments[i]); err != nil {
			log.Fatalf("failed to create assignment: %v", err)
		}
	}

	// Seed documentation entries.
	docEntries := []models.DocumentationEntry{
		{ChildID: 1, TeacherID: 1, CategoryID: 1, ObservationDescription: "Anna zeigt große Hilfsbereitschaft gegenüber anderen Kindern. Sie hilft beim Aufräumen und tröstet weinende Kinder.", ObservationDate: parseDate("2024-01-15"), IsApproved: true, ApprovedByUserID: intPtr(1)},
		{ChildID: 1, TeacherID: 1, CategoryID: 2, ObservationDescription: "Anna verwendet komplexe Sätze und erzählt zusammenhängende Geschichten. Ihr Wortschatz erweitert sich täglich.", ObservationDate: parseDate("2024-02-10"), IsApproved: true, ApprovedByUserID: intPtr(1)},
		{ChildID: 2, TeacherID: 1, CategoryID: 2, ObservationDescription: "Liam macht große Fortschritte in der deutschen Sprache. Er kommuniziert zunehmend auf Deutsch mit anderen Kindern.", ObservationDate: parseDate("2024-01-20"), IsApproved: true, ApprovedByUserID: intPtr(1)},
		{ChildID: 3, TeacherID: 2, CategoryID: 4, ObservationDescription: "Ben löst Puzzles mit 50+ Teilen selbständig und zeigt dabei große Ausdauer und logisches Denken.", ObservationDate: parseDate("2024-02-05"), IsApproved: true, ApprovedByUserID: intPtr(2)},
		{ChildID: 4, TeacherID: 2, CategoryID: 3, ObservationDescription: "Noah zeigt ausgezeichnete Feinmotorik beim Basteln und kann bereits seinen Namen schreiben.", ObservationDate: parseDate("2024-01-25"), IsApproved: false, ApprovedByUserID: nil},
		{ChildID: 5, TeacherID: 3, CategoryID: 6, ObservationDescription: "Mia reguliert ihre Emotionen sehr gut und kann Konflikte verbal lösen, anstatt zu weinen oder zu schreien.", ObservationDate: parseDate("2024-02-12"), IsApproved: true, ApprovedByUserID: intPtr(3)},
		{ChildID: 6, TeacherID: 3, CategoryID: 1, ObservationDescription: "Lucas integriert sich gut in die Gruppe und hat bereits enge Freundschaften entwickelt.", ObservationDate: parseDate("2024-01-30"), IsApproved: true, ApprovedByUserID: intPtr(3)},
		{ChildID: 7, TeacherID: 4, CategoryID: 5, ObservationDescription: "Charlotte zeigt große Kreativität beim Malen und Basteln. Ihre Kunstwerke sind sehr detailreich und fantasievoll.", ObservationDate: parseDate("2024-02-08"), IsApproved: true, ApprovedByUserID: intPtr(4)},
		{ChildID: 8, TeacherID: 4, CategoryID: 3, ObservationDescription: "Oliver turnt gerne und zeigt gute Koordination beim Klettern und Balancieren.", ObservationDate: parseDate("2024-02-01"), IsApproved: false, ApprovedByUserID: nil},
		{ChildID: 1, TeacherID: 1, CategoryID: 4, ObservationDescription: "Anna zeigt Interesse an mathematischen Konzepten und kann bis 20 zählen.", ObservationDate: parseDate("2024-02-20"), IsApproved: false, ApprovedByUserID: nil},
	}

	for i := range docEntries {
		if _, err := dal.DocumentationEntries.Create(&docEntries[i]); err != nil {
			log.Fatalf("failed to create documentation entry: %v", err)
		}
	}

	fmt.Println("Database seeded successfully")
}

func intPtr(i int) *int { return &i }
