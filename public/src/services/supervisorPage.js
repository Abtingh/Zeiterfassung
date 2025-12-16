/**
 * Supervisor Zeit Genehmigen Page Controller
 * Handles UI interactions for approving/rejecting time entries
 */

let currentSubmission = null;
let currentWeek = null;
let currentYear = null;

document.addEventListener('DOMContentLoaded', async () => {
    await initializePage();
    setupEventListeners();
});

/**
 * Initialize the page - load pending submissions and populate dropdown
 */
async function initializePage() {
    try {
        console.log('Initializing supervisor page...');
        
        // Load pending submissions (excludes 'offen' status)
        const submissions = await supervisorService.getPendingSubmissions();
        
        console.log('Received submissions:', submissions);
        console.log('Number of submissions:', submissions ? submissions.length : 'null');
        
        if (!submissions || submissions.length === 0) {
            showMessage('Keine ausstehenden Einträge vorhanden.', 'info');
            disableControls();
            return;
        }

        // Populate student dropdown
        populateStudentDropdown();
        
        // Check for URL parameters (from home page navigation)
        const urlParams = new URLSearchParams(window.location.search);
        const studentIdParam = urlParams.get('student_id');
        const yearParam = urlParams.get('year');
        const monthParam = urlParams.get('month');
        const weekParam = urlParams.get('week');
        
        console.log('URL params:', { studentIdParam, yearParam, monthParam, weekParam });
        
        if (studentIdParam && yearParam && monthParam && weekParam) {
            // Calculate global week number from month and week params
            const month0 = parseInt(monthParam) - 1; // Convert to 0-indexed
            const weekIndex = parseInt(weekParam);
            const globalWeekNumber = (month0 * 5) + weekIndex;
            
            console.log('Looking for student:', studentIdParam, 'week:', globalWeekNumber, 'year:', yearParam);
            
            // Set student dropdown
            document.getElementById('studentSelect').value = studentIdParam;
            
            // Find matching submission
            const targetSubmission = submissions.find(s => 
                s.user_id === parseInt(studentIdParam) && 
                s.week_number === globalWeekNumber && 
                s.year === parseInt(yearParam)
            );
            
            if (targetSubmission) {
                console.log('Found target submission:', targetSubmission);
                currentSubmission = targetSubmission;
                currentWeek = targetSubmission.week_number;
                currentYear = targetSubmission.year;
                
                updateWeekDisplay();
                updateStatusDisplay();
                await loadTimeEntriesForSubmission();
                return;
            } else {
                console.log('No matching submission found for params, falling back to first submission');
            }
        }
        
        // Default: Select first student
        const students = supervisorService.getUniqueStudents();
        console.log('Unique students:', students);
        
        if (students.length > 0) {
            document.getElementById('studentSelect').value = students[0].id;
            await onStudentChange(students[0].id);
        }
    } catch (error) {
        console.error('Error initializing page:', error);
        showMessage('Fehler beim Laden der Daten.', 'error');
    }
}

/**
 * Populate student dropdown with students who have pending submissions
 */
function populateStudentDropdown() {
    const select = document.getElementById('studentSelect');
    const students = supervisorService.getUniqueStudents();
    
    // Clear existing options except placeholder
    select.innerHTML = '<option value="" disabled>Student auswählen</option>';
    
    students.forEach(student => {
        const option = document.createElement('option');
        option.value = student.id;
        option.textContent = student.name || student.email;
        select.appendChild(option);
    });
}

/**
 * Handle student selection change
 * @param {number} userId - Selected user ID
 */
async function onStudentChange(userId) {
    const submissions = supervisorService.getSubmissionsForStudent(parseInt(userId));
    
    if (submissions.length === 0) {
        showMessage('Keine ausstehenden Einträge für diesen Studenten.', 'info');
        return;
    }
    
    // Sort submissions by year and week (ascending) to match navigation order
    const sortedSubmissions = [...submissions].sort((a, b) => {
        if (a.year !== b.year) return a.year - b.year;
        return a.week_number - b.week_number;
    });
    
    // Calculate current week's global week number
    const now = new Date();
    const currentMonth0 = now.getMonth();
    const currentYearNow = now.getFullYear();
    
    // Find which week of the month we're in using fridaysInMonth
    let currentGlobalWeek = null;
    if (typeof fridaysInMonth === 'function') {
        const fridays = fridaysInMonth(currentYearNow, currentMonth0);
        for (let i = 0; i < fridays.length; i++) {
            const friday = fridays[i];
            const monday = new Date(friday);
            monday.setDate(friday.getDate() - 4);
            const sunday = new Date(friday);
            sunday.setDate(friday.getDate() + 2);
            
            if (now >= monday && now <= sunday) {
                currentGlobalWeek = (currentMonth0 * 5) + (i + 1);
                break;
            }
        }
    }
    
    console.log('Current global week:', currentGlobalWeek);
    
    // Try to find submission for current week
    let selectedSubmission = null;
    if (currentGlobalWeek) {
        selectedSubmission = sortedSubmissions.find(s => 
            s.week_number === currentGlobalWeek && s.year === currentYearNow
        );
    }
    
    // If no current week submission, use the most recent one
    if (!selectedSubmission) {
        selectedSubmission = sortedSubmissions[sortedSubmissions.length - 1];
    }
    
    currentSubmission = selectedSubmission;
    currentWeek = currentSubmission.week_number;
    currentYear = currentSubmission.year;
    
    console.log('Selected submission:', currentSubmission.id, 'Week:', currentWeek, 'Year:', currentYear);
    
    // Update UI
    updateWeekDisplay();
    updateStatusDisplay();
    
    // Load time entries for this submission
    await loadTimeEntriesForSubmission();
}

/**
 * Update week display in header
 */
function updateWeekDisplay() {
    const weekHolder = document.querySelector('#weekHolder p');
    const dateHolder = document.getElementById('date');
    
    if (currentWeek && currentYear) {
        // Convert global week number back to display week
        // Global week = (month * 5) + weekInMonth
        // weekInMonth ranges from 1-5
        const weekInMonth = ((currentWeek - 1) % 5) + 1;
        
        weekHolder.textContent = `${weekInMonth}.Woche`;
        
        // Calculate week date range based on actual dates
        const weekDates = getWeekDateRangeFromGlobal(currentWeek, currentYear);
        dateHolder.textContent = `${weekDates.start} - ${weekDates.end}`;
    }
}

/**
 * Update status display
 */
function updateStatusDisplay() {
    const statusText = document.getElementById('statusText');
    const editBtn = document.getElementById('editBtn');
    const submitBtn = document.getElementById('submitBtn');
    
    if (currentSubmission) {
        const statusMap = {
            'offen': 'Offen',
            'entwurf': 'Entwurf',
            'gesendet': 'Gesendet',
            'bestaetigt': 'Bestätigt',
            'korrektur': 'Korrektur',
            'erledigt': 'Erledigt'
        };
        statusText.textContent = statusMap[currentSubmission.status] || currentSubmission.status;
        
        // Add color based on status
        statusText.className = '';
        statusText.classList.add(`status-${currentSubmission.status}`);
        
        // Enable/disable buttons based on status
        // Only 'gesendet' status should have both buttons enabled
        if (currentSubmission.status === 'gesendet') {
            editBtn.disabled = false;
            submitBtn.disabled = false;
            editBtn.style.opacity = '1';
            submitBtn.style.opacity = '1';
            editBtn.style.cursor = 'pointer';
            submitBtn.style.cursor = 'pointer';
        } else {
            editBtn.disabled = true;
            submitBtn.disabled = true;
            editBtn.style.opacity = '0.5';
            submitBtn.style.opacity = '0.5';
            editBtn.style.cursor = 'not-allowed';
            submitBtn.style.cursor = 'not-allowed';
        }
    }
}

/**
 * Load time entries for the current submission
 */
async function loadTimeEntriesForSubmission() {
    if (!currentSubmission) {
        console.log('No currentSubmission set');
        return;
    }
    
    try {
        console.log('Loading time entries for submission:', currentSubmission.id);
        
        // Use supervisorService to get time entries for this specific submission
        const entries = await supervisorService.getTimeEntriesForSubmission(currentSubmission.id);
        console.log('Received entries:', entries);
        
        if (entries && entries.length > 0) {
            populateTableWithEntries(entries);
        } else {
            console.log('No entries found for this submission');
            clearTable();
        }
    } catch (error) {
        console.error('Error loading time entries:', error);
    }
}

/**
 * Update table dates based on current week
 */
function updateTableDates() {
    if (!currentWeek || !currentYear) return;
    
    // Decode global week to get month and week in month
    const month0 = Math.floor((currentWeek - 1) / 5);
    const weekInMonth = ((currentWeek - 1) % 5) + 1;
    
    // Get the Friday for this week using fridaysInMonth
    if (typeof fridaysInMonth === 'function') {
        const fridays = fridaysInMonth(currentYear, month0);
        if (fridays.length >= weekInMonth) {
            const friday = fridays[weekInMonth - 1];
            const monday = new Date(friday);
            monday.setDate(friday.getDate() - 4);
            
            // Update each day's date in the table
            const dayClasses = ['first', 'second', 'third', 'fourth', 'fifth', 'sixth', 'seventh'];
            dayClasses.forEach((cls, i) => {
                const dateElement = document.querySelector(`.${cls}.day`);
                if (dateElement) {
                    const dayDate = new Date(monday);
                    dayDate.setDate(monday.getDate() + i);
                    const formatted = `${dayDate.getDate().toString().padStart(2, '0')}.${(dayDate.getMonth() + 1).toString().padStart(2, '0')}`;
                    dateElement.textContent = formatted;
                }
            });
        }
    }
}

/**
 * Clear the time entry table
 */
function clearTable() {
    const rows = document.querySelectorAll('#timeTable tbody tr:not(.total-row)');
    rows.forEach(row => {
        const timeInputs = row.querySelectorAll('.time-input');
        const pauseInput = row.querySelector('.pause-input');
        const notesInput = row.querySelector('.notes-input');
        const calcHours = row.querySelector('.calculated-hours');
        const workHours = row.querySelector('.work-hours');
        
        if (timeInputs[0]) timeInputs[0].value = '';
        if (timeInputs[1]) timeInputs[1].value = '';
        if (pauseInput) pauseInput.value = '';
        if (notesInput) notesInput.value = '';
        if (calcHours) calcHours.textContent = '-';
        if (workHours) workHours.textContent = '-';
    });
}

/**
 * Format time from microseconds (pgtype.Time format)
 */
function formatTimeFromMicroseconds(timeValue) {
    if (!timeValue) return '';
    
    // Handle pgtype.Time format
    if (typeof timeValue === 'object' && timeValue.Microseconds !== undefined) {
        const microseconds = timeValue.Microseconds;
        const hours = Math.floor(microseconds / 3600000000);
        const minutes = Math.floor((microseconds % 3600000000) / 60000000);
        return `${String(hours).padStart(2, '0')}:${String(minutes).padStart(2, '0')}`;
    }
    
    // If it's already a string
    if (typeof timeValue === 'string') {
        return timeValue.substring(0, 5);
    }
    
    return '';
}

/**
 * Calculate hours difference between two times
 */
function calculateHoursDiff(startTime, endTime) {
    if (!startTime || !endTime) return 0;
    
    const [startH, startM] = startTime.split(':').map(Number);
    const [endH, endM] = endTime.split(':').map(Number);
    
    const startMinutes = startH * 60 + startM;
    const endMinutes = endH * 60 + endM;
    
    return (endMinutes - startMinutes) / 60;
}

/**
 * Populate table with time entry data
 * @param {Array} entries - Array of time entries
 */
function populateTableWithEntries(entries) {
    const rows = document.querySelectorAll('#timeTable tbody tr:not(.total-row)');
    
    // Clear table first
    clearTable();
    
    // Update table dates based on current week
    updateTableDates();
    
    // Populate with data
    let totalHours = 0;
    console.log('Populating table with entries:', entries);
    console.log('Number of table rows:', rows.length);
    
    // Create a map of day-of-week to entry
    // We need to match entry_date to the correct row (Mo=0, Di=1, etc.)
    const dayLabels = ['Mo', 'Di', 'Mi', 'Do', 'Fr', 'Sa', 'So'];
    
    entries.forEach((entry) => {
        console.log('Processing entry:', entry);
        
        // Get the day of week from entry_date
        let dayIndex = -1;
        if (entry.entry_date) {
            // Handle pgtype.Date format
            let dateStr = entry.entry_date;
            if (typeof entry.entry_date === 'object' && entry.entry_date.Time) {
                dateStr = entry.entry_date.Time;
            }
            
            const entryDate = new Date(dateStr);
            // getDay() returns 0 for Sunday, 1 for Monday, etc.
            // We need: Monday=0, Tuesday=1, ..., Sunday=6
            dayIndex = (entryDate.getDay() + 6) % 7;
            console.log('Entry date:', dateStr, 'Day index:', dayIndex);
        }
        
        if (dayIndex < 0 || dayIndex >= rows.length) {
            console.log('Invalid day index:', dayIndex);
            return;
        }
        
        const row = rows[dayIndex];
        const timeInputs = row.querySelectorAll('.time-input');
        const pauseInput = row.querySelector('.pause-input');
        const notesInput = row.querySelector('.notes-input');
        const calcHours = row.querySelector('.calculated-hours');
        const workHours = row.querySelector('.work-hours');
        
        // Format time values
        if (timeInputs[0] && entry.start_time) {
            const startTime = formatTimeFromMicroseconds(entry.start_time);
            console.log('Setting start time:', startTime, 'for row', dayIndex);
            timeInputs[0].value = startTime;
        }
        if (timeInputs[1] && entry.end_time) {
            const endTime = formatTimeFromMicroseconds(entry.end_time);
            console.log('Setting end time:', endTime, 'for row', dayIndex);
            timeInputs[1].value = endTime;
        }
        
        if (pauseInput) {
            const breakMin = entry.break_min?.Int32 || entry.break_min || 0;
            pauseInput.value = breakMin;
        }
        
        if (notesInput && entry.note) {
            const noteText = entry.note?.String || entry.note || '';
            notesInput.value = noteText;
        }
        
        // Duration
        let duration = 0;
        if (entry.duration_h) {
            if (typeof entry.duration_h === 'object' && entry.duration_h.String) {
                duration = parseFloat(entry.duration_h.String) || 0;
            } else {
                duration = parseFloat(entry.duration_h) || 0;
            }
        }
        
        if (workHours) {
            workHours.textContent = duration.toFixed(2);
            totalHours += duration;
        }
        
        // Calculate hours if we have start and end time
        if (entry.start_time && entry.end_time && calcHours) {
            const startTime = formatTimeFromMicroseconds(entry.start_time);
            const endTime = formatTimeFromMicroseconds(entry.end_time);
            const hours = calculateHoursDiff(startTime, endTime);
            calcHours.textContent = hours.toFixed(2);
        }
    });
    
    // Update total
    const totalElement = document.getElementById('totalWorkHours');
    if (totalElement) {
        totalElement.textContent = totalHours.toFixed(2);
    }
    
    // Disable inputs for supervisor view (read-only)
    disableInputs();
}

/**
 * Disable all inputs (supervisor view is read-only)
 */
function disableInputs() {
    document.querySelectorAll('#timeTable input').forEach(input => {
        input.disabled = true;
    });
}

/**
 * Disable controls when no data
 */
function disableControls() {
    document.getElementById('submitBtn').disabled = true;
    document.getElementById('editBtn').disabled = true;
}

/**
 * Setup event listeners
 */
function setupEventListeners() {
    // Student dropdown change
    document.getElementById('studentSelect').addEventListener('change', (e) => {
        onStudentChange(e.target.value);
    });
    
    // Approve button (Freigeben)
    document.getElementById('submitBtn').addEventListener('click', async () => {
        if (!currentSubmission) {
            showMessage('Bitte wählen Sie einen Eintrag aus.', 'warning');
            return;
        }
        
        if (confirm('Möchten Sie diesen Zeiteintrag freigeben?')) {
            try {
                await supervisorService.approveSubmission(currentSubmission.id);
                showMessage('Zeiteintrag wurde freigegeben!', 'success');
                
                // Reload page to refresh data
                setTimeout(() => location.reload(), 1500);
            } catch (error) {
                showMessage('Fehler beim Freigeben: ' + error.message, 'error');
            }
        }
    });
    
    // Reject button (korrigieren)
    document.getElementById('editBtn').addEventListener('click', async () => {
        if (!currentSubmission) {
            showMessage('Bitte wählen Sie einen Eintrag aus.', 'warning');
            return;
        }
        
        const reason = prompt('Grund für die Korrekturanforderung (optional):');
        
        if (confirm('Möchten Sie diesen Zeiteintrag zur Korrektur zurücksenden?')) {
            try {
                await supervisorService.rejectSubmission(currentSubmission.id, reason || '');
                showMessage('Zeiteintrag wurde zur Korrektur zurückgesendet.', 'success');
                
                // Reload page to refresh data
                setTimeout(() => location.reload(), 1500);
            } catch (error) {
                showMessage('Fehler beim Ablehnen: ' + error.message, 'error');
            }
        }
    });
    
    // Week navigation (prev/next)
    document.getElementById('prevBtn').addEventListener('click', () => {
        navigateWeek(-1);
    });
    
    document.getElementById('nextBtn').addEventListener('click', () => {
        navigateWeek(1);
    });
}

/**
 * Navigate to previous/next week
 * @param {number} direction - -1 for previous, 1 for next
 */
function navigateWeek(direction) {
    if (!currentSubmission) {
        console.log('No current submission');
        return;
    }
    
    const userId = currentSubmission.user_id;
    const submissions = supervisorService.getSubmissionsForStudent(userId);
    
    console.log('All submissions for user:', submissions.length);
    
    // Sort submissions by year and week number (ascending)
    const sortedSubmissions = [...submissions].sort((a, b) => {
        if (a.year !== b.year) return a.year - b.year;
        return a.week_number - b.week_number;
    });
    
    console.log('Sorted submissions:', sortedSubmissions.map(s => ({ week: s.week_number, year: s.year, status: s.status, id: s.id.substring(0,8) })));
    console.log('Current submission id:', currentSubmission.id.substring(0,8), 'week:', currentWeek);
    
    // Find current index by matching submission ID (most reliable)
    const currentIndex = sortedSubmissions.findIndex(s => s.id === currentSubmission.id);
    
    console.log('Current index:', currentIndex, 'Direction:', direction);
    
    if (currentIndex === -1) {
        console.error('Current submission not found in sorted list!');
        return;
    }
    
    // Calculate target index
    const targetIndex = currentIndex + direction;
    
    console.log('Target index:', targetIndex, 'Max:', sortedSubmissions.length - 1);
    
    if (targetIndex >= 0 && targetIndex < sortedSubmissions.length) {
        const targetSubmission = sortedSubmissions[targetIndex];
        console.log('Navigating to:', targetSubmission.id.substring(0,8), 'week:', targetSubmission.week_number, 'status:', targetSubmission.status);
        
        currentSubmission = targetSubmission;
        currentWeek = targetSubmission.week_number;
        currentYear = targetSubmission.year;
        
        updateWeekDisplay();
        updateStatusDisplay();
        loadTimeEntriesForSubmission();
    } else {
        console.log('Cannot navigate - at boundary. targetIndex:', targetIndex);
        // Do NOT change anything if we can't navigate
    }
}

/**
 * Calculate hours between two times
 * @param {string} start - Start time (HH:MM)
 * @param {string} end - End time (HH:MM)
 * @returns {number} Hours difference
 */
function calculateHours(start, end) {
    const [startH, startM] = start.split(':').map(Number);
    const [endH, endM] = end.split(':').map(Number);
    
    const startMinutes = startH * 60 + startM;
    const endMinutes = endH * 60 + endM;
    
    return (endMinutes - startMinutes) / 60;
}

/**
 * Get week date range from global week number
 * Global week = (month * 5) + weekInMonth
 * Uses the same business rules as fridaysInMonth()
 * 
 * @param {number} globalWeek - The global week number
 * @param {number} year - Year
 * @returns {Object} Start and end dates formatted
 */
function getWeekDateRangeFromGlobal(globalWeek, year) {
    // Decode global week: globalWeek = (month * 5) + weekInMonth
    // So: month = floor((globalWeek - 1) / 5), weekInMonth = ((globalWeek - 1) % 5) + 1
    const month0 = Math.floor((globalWeek - 1) / 5);
    const weekInMonth = ((globalWeek - 1) % 5) + 1;
    
    // Use the same fridaysInMonth logic to find the exact Friday
    if (typeof fridaysInMonth === 'function') {
        const fridays = fridaysInMonth(year, month0);
        if (fridays.length >= weekInMonth) {
            const friday = fridays[weekInMonth - 1];
            const monday = new Date(friday);
            monday.setDate(friday.getDate() - 4);
            const sunday = new Date(friday);
            sunday.setDate(friday.getDate() + 2);
            
            const formatDate = (d) => `${d.getDate().toString().padStart(2, '0')}.${(d.getMonth() + 1).toString().padStart(2, '0')}`;
            
            return {
                start: formatDate(monday),
                end: formatDate(sunday)
            };
        }
    }
    
    // Fallback to simple calculation
    return getWeekDateRange(globalWeek, year);
}

/**
 * Get week date range
 * @param {number} week - Week number
 * @param {number} year - Year
 * @returns {Object} Start and end dates formatted
 */
function getWeekDateRange(week, year) {
    const simple = new Date(year, 0, 1 + (week - 1) * 7);
    const dow = simple.getDay();
    const monday = new Date(simple);
    if (dow <= 4) {
        monday.setDate(simple.getDate() - simple.getDay() + 1);
    } else {
        monday.setDate(simple.getDate() + 8 - simple.getDay());
    }
    
    const sunday = new Date(monday);
    sunday.setDate(monday.getDate() + 6);
    
    const formatDate = (d) => `${d.getDate().toString().padStart(2, '0')}.${(d.getMonth() + 1).toString().padStart(2, '0')}`;
    
    return {
        start: formatDate(monday),
        end: formatDate(sunday)
    };
}

/**
 * Show message to user
 * @param {string} message - Message text
 * @param {string} type - Message type: 'success', 'error', 'warning', 'info'
 */
function showMessage(message, type = 'info') {
    // Create message element if it doesn't exist
    let msgBox = document.getElementById('messageBox');
    if (!msgBox) {
        msgBox = document.createElement('div');
        msgBox.id = 'messageBox';
        msgBox.style.cssText = `
            position: fixed;
            top: 20px;
            right: 20px;
            padding: 15px 25px;
            border-radius: 8px;
            z-index: 1000;
            font-weight: 500;
            box-shadow: 0 4px 12px rgba(0,0,0,0.15);
            transition: opacity 0.3s;
        `;
        document.body.appendChild(msgBox);
    }
    
    // Set colors based on type
    const colors = {
        success: { bg: '#d4edda', text: '#155724', border: '#c3e6cb' },
        error: { bg: '#f8d7da', text: '#721c24', border: '#f5c6cb' },
        warning: { bg: '#fff3cd', text: '#856404', border: '#ffeeba' },
        info: { bg: '#d1ecf1', text: '#0c5460', border: '#bee5eb' }
    };
    
    const color = colors[type] || colors.info;
    msgBox.style.backgroundColor = color.bg;
    msgBox.style.color = color.text;
    msgBox.style.border = `1px solid ${color.border}`;
    
    msgBox.textContent = message;
    msgBox.style.opacity = '1';
    msgBox.style.display = 'block';
    
    // Auto-hide after 3 seconds
    setTimeout(() => {
        msgBox.style.opacity = '0';
        setTimeout(() => msgBox.style.display = 'none', 300);
    }, 3000);
}
