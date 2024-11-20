package utils

// Helper function to check valid days
func IsValidDayOfWeek(dayOfWeek string) bool {
	validDays := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	for _, day := range validDays {
		if day == dayOfWeek {
			return true
		}
	}
	return false
}
