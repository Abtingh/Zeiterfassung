class TimeEntryService {
    constructor() {
        this.baseUrl = '/api/time-entries';
    }

    // Get current week number and year
    getWeekInfo(date = new Date()) {
        const startDate = new Date(date.getFullYear(), 0, 1);
        const days = Math.floor((date - startDate) / (24 * 60 * 60 * 1000));
        const weekNumber = Math.ceil((days + startDate.getDay() + 1) / 7);
        return {
            week: weekNumber,
            year: date.getFullYear()
        };
    }

    // Submit weekly time entries
    async submitWeek(weekData) {
        try {
            console.log('Submitting data:', weekData);

            const response = await fetch(`${this.baseUrl}/submit`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
                body: JSON.stringify(weekData)
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.error || 'Failed to submit time entries');
            }

            return data;
        } catch (error) {
            console.error('Error submitting time entries:', error);
            throw error;
        }
    }

    // Load time entries for a specific week
    async loadWeek(week, year) {
        try {
            const response = await fetch(`${this.baseUrl}/week?week=${week}&year=${year}`, {
                method: 'GET',
                credentials: 'include'
            });

            const data = await response.json();

            if (!response.ok) {
                throw new Error(data.error || 'Failed to load time entries');
            }

            return data;
        } catch (error) {
            console.error('Error loading time entries:', error);
            throw error;
        }
    }

    // Convert table data to API format
    convertToAPIFormat(tableData) {
        const entries = tableData.weekData.map(dayEntry => ({
            date: this.parseDate(dayEntry.date, tableData.year),
            startTime: dayEntry.from || "",
            endTime: dayEntry.to || "",
            breakMin: dayEntry.pause,
            durationH: dayEntry.workHours,
            note: dayEntry.notes || ""
        }));

        return {
            weekNumber: tableData.weekNumber,
            year: tableData.year,
            entries: entries
        };
    }

    // Parse DD.MM format to YYYY-MM-DD
    parseDate(dateStr, year) {
        const [day, month] = dateStr.split('.');
        return `${year}-${month.padStart(2, '0')}-${day.padStart(2, '0')}`;
    }

    // Populate table with loaded data
    populateTable(entries) {
        const rows = document.querySelectorAll('#timeTable tbody tr:not(.total-row)');

        // Clear all inputs first
        rows.forEach(row => {
            const timeInputs = row.querySelectorAll('.time-input:not([disabled])');
            const pauseInput = row.querySelector('.pause-input');
            const notesInput = row.querySelector('.notes-input:not([disabled])');

            if (timeInputs.length > 0) {
                timeInputs[0].value = '';
                timeInputs[1].value = '';
            }
            if (pauseInput) pauseInput.value = '';
            if (notesInput) notesInput.value = '';
        });

        // Populate with loaded data
        entries.forEach(entry => {
            console.log('Processing entry:', entry);
            
            // Parse the date - it comes as a string "YYYY-MM-DD" from the database
            const entryDate = new Date(entry.entry_date);
            console.log('Entry date:', entryDate, 'Day index:', entryDate.getDay());
            
            const dayIndex = entryDate.getDay();
            const rowIndex = dayIndex === 0 ? 6 : dayIndex - 1;
            const row = rows[rowIndex];

            if (row && !row.querySelector('.time-input').disabled) {
                const timeInputs = row.querySelectorAll('.time-input');
                const pauseInput = row.querySelector('.pause-input');
                const notesInput = row.querySelector('.notes-input');

                // Populate start time - handle both pgtype.Time format and string format
                if (entry.start_time) {
                    let timeStr = '';
                    if (typeof entry.start_time === 'string') {
                        // String format "HH:MM:SS"
                        timeStr = entry.start_time.substring(0, 5); // Get "HH:MM"
                    } else if (entry.start_time.Valid && entry.start_time.Microseconds) {
                        // pgtype.Time format
                        const microseconds = entry.start_time.Microseconds;
                        const totalSeconds = Math.floor(microseconds / 1000000);
                        const hours = Math.floor(totalSeconds / 3600);
                        const minutes = Math.floor((totalSeconds % 3600) / 60);
                        timeStr = `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}`;
                    }
                    if (timeStr) {
                        console.log('Setting start time:', timeStr);
                        timeInputs[0].value = timeStr;
                    }
                }

                // Populate end time - handle both pgtype.Time format and string format
                if (entry.end_time) {
                    let timeStr = '';
                    if (typeof entry.end_time === 'string') {
                        // String format "HH:MM:SS"
                        timeStr = entry.end_time.substring(0, 5); // Get "HH:MM"
                    } else if (entry.end_time.Valid && entry.end_time.Microseconds) {
                        // pgtype.Time format
                        const microseconds = entry.end_time.Microseconds;
                        const totalSeconds = Math.floor(microseconds / 1000000);
                        const hours = Math.floor(totalSeconds / 3600);
                        const minutes = Math.floor((totalSeconds % 3600) / 60);
                        timeStr = `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}`;
                    }
                    if (timeStr) {
                        console.log('Setting end time:', timeStr);
                        timeInputs[1].value = timeStr;
                    }
                }

                // Populate break - handle both formats
                if (entry.break_min !== undefined && entry.break_min !== null) {
                    let breakValue = 0;
                    if (typeof entry.break_min === 'number') {
                        breakValue = entry.break_min;
                    } else if (typeof entry.break_min === 'object') {
                        // Handle pgtype format - could be Int32, Int64, or Float64
                        if (entry.break_min.Valid) {
                            breakValue = entry.break_min.Int32 || entry.break_min.Int64 || entry.break_min.Float64 || 0;
                        }
                    }
                    console.log('Setting break:', breakValue, 'from:', entry.break_min);
                    if (pauseInput) {
                        pauseInput.value = breakValue;
                    }
                }

                // Populate note
                if (entry.note && entry.note.Valid) {
                    notesInput.value = entry.note.String;
                }
            }
        });

        // Trigger recalculation for all rows AFTER all data is populated
        setTimeout(() => {
            rows.forEach(row => {
                const timeInputs = row.querySelectorAll('.time-input:not([disabled])');
                if (timeInputs.length > 0 && timeInputs[0].value) {
                    timeInputs[0].dispatchEvent(new Event('change', { bubbles: true }));
                }
            });
        }, 100);
    }
}

// Initialize service
const timeEntryService = new TimeEntryService();
window.timeEntryService = timeEntryService;