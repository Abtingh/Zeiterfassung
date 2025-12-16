/**
 * Accounting Zeiten Übersicht Page Controller
 * Handles UI interactions for viewing and processing approved time entries
 */

let currentSubmission = null;
let currentStudent = null;
let currentWeek = null;
let currentYear = null;

document.addEventListener('DOMContentLoaded', async () => {
    await initializePage();
    setupEventListeners();
});

/**
 * Initialize the page - load students and approved submissions
 */
async function initializePage() {
    try {
        // Load all students
        await accountingService.getAllStudents();
        
        // Load approved submissions
        const submissions = await accountingService.getApprovedSubmissions();
        
        // Populate student dropdown
        populateStudentDropdown();
        
        if (submissions.length === 0) {
            showMessage('Keine genehmigten Einträge vorhanden.', 'info');
        }
        
        // Select first student with approved submissions by default
        const studentsWithSubmissions = getStudentsWithSubmissions();
        if (studentsWithSubmissions.length > 0) {
            document.getElementById('studentSelect').value = studentsWithSubmissions[0].id;
            await onStudentChange(studentsWithSubmissions[0].id);
        }
    } catch (error) {
        console.error('Error initializing page:', error);
        showMessage('Fehler beim Laden der Daten.', 'error');
    }
}

/**
 * Get students who have approved submissions
 */
function getStudentsWithSubmissions() {
    const studentIds = new Set(accountingService.approvedSubmissions.map(s => s.user_id));
    return accountingService.allStudents.filter(s => studentIds.has(s.id));
}

/**
 * Populate student dropdown with all students
 */
function populateStudentDropdown() {
    const select = document.getElementById('studentSelect');
    const students = accountingService.allStudents;
    
    // Clear existing options except placeholder
    select.innerHTML = '<option value="" disabled>Student auswählen</option>';
    
    students.forEach(student => {
        const option = document.createElement('option');
        option.value = student.id;
        const firstName = student.first_name?.String || student.first_name || '';
        const lastName = student.last_name?.String || student.last_name || '';
        option.textContent = `${firstName} ${lastName}`.trim() || student.email;
        select.appendChild(option);
    });
}

/**
 * Handle student selection change
 * @param {number} userId - Selected user ID
 */
async function onStudentChange(userId) {
    const studentId = parseInt(userId);
    currentStudent = accountingService.getStudentById(studentId);
    
    // Update team display
    updateTeamDisplay();
    
    // Get ALL submissions for this student (any status)
    try {
        const submissions = await accountingService.getAllSubmissionsForStudent(studentId);
        
        if (!submissions || submissions.length === 0) {
            showMessage('Keine Einträge für diesen Studenten vorhanden.', 'info');
            clearTable();
            disableSubmitButton();
            updateWeekDisplay();
            updateTableDates();
            updateStatusDisplay();
            return;
        }
        
        // Sort submissions by year and week (ascending) to match navigation order
        const sortedSubmissions = [...submissions].sort((a, b) => {
            if (a.year !== b.year) return a.year - b.year;
            return a.week_number - b.week_number;
        });
        
        // Store sorted submissions for navigation
        accountingService.studentSubmissions = sortedSubmissions;
        
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
        updateTableDates();
        updateStatusDisplay();
        
        // Load time entries for this submission
        await loadTimeEntriesForSubmission();
        
        // Only enable submit button if status is 'bestaetigt' (approved)
        if (currentSubmission.status === 'bestaetigt') {
            enableSubmitButton();
        } else {
            disableSubmitButton();
        }
    } catch (error) {
        console.error('Error loading student submissions:', error);
        showMessage('Fehler beim Laden der Daten.', 'error');
    }
}

/**
 * Update team display
 */
function updateTeamDisplay() {
    const teamHolder = document.getElementById('teamHolder');
    if (teamHolder && currentStudent) {
        const teamName = currentStudent.team_name?.String || currentStudent.team_name || 'Kein Team';
        teamHolder.innerHTML = `<p>Team: <span id="teamName">${teamName}</span></p>`;
    }
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
    if (currentSubmission) {
        const statusMap = {
            'offen': 'Offen',
            'gesendet': 'Gesendet',
            'bestaetigt': 'Bestätigt',
            'korrektur': 'Korrektur',
            'erledigt': 'Erledigt'
        };
        statusText.textContent = statusMap[currentSubmission.status] || currentSubmission.status;
        
        // Add color based on status
        statusText.className = '';
        statusText.classList.add(`status-${currentSubmission.status}`);
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
 * Load time entries for the current submission
 */
async function loadTimeEntriesForSubmission() {
    if (!currentSubmission) return;
    
    try {
        console.log('Loading time entries for submission:', currentSubmission.id);
        
        const entries = await accountingService.getTimeEntriesForSubmission(currentSubmission.id);
        
        console.log('Received entries:', entries);
        
        if (entries && entries.length > 0) {
            populateTableWithEntries(entries);
        } else {
            clearTable();
        }
    } catch (error) {
        console.error('Error loading time entries:', error);
        clearTable();
    }
}

/**
 * Populate table with time entry data
 * @param {Array} entries - Array of time entries
 */
function populateTableWithEntries(entries) {
    const rows = document.querySelectorAll('#timeTable tbody tr:not(.total-row)');
    
    // Clear table first
    clearTable();
    
    // Populate with data
    let totalHours = 0;
    entries.forEach((entry, index) => {
        if (index < rows.length) {
            const row = rows[index];
            const timeInputs = row.querySelectorAll('.time-input');
            const pauseCell = row.querySelector('.pause-input') || row.querySelector('.pause-time');
            const notesInput = row.querySelector('.notes-input');
            const calcHours = row.querySelector('.calculated-hours');
            const workHours = row.querySelector('.work-hours');
            
            // Format time values
            if (timeInputs[0] && entry.start_time) {
                const startTime = formatTimeFromMicroseconds(entry.start_time);
                timeInputs[0].value = startTime;
            }
            if (timeInputs[1] && entry.end_time) {
                const endTime = formatTimeFromMicroseconds(entry.end_time);
                timeInputs[1].value = endTime;
            }
            
            if (pauseCell) {
                const breakMin = entry.break_min?.Int32 || entry.break_min || 0;
                if (pauseCell.tagName === 'INPUT') {
                    pauseCell.value = breakMin;
                } else {
                    pauseCell.textContent = breakMin;
                }
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
        }
    });
    
    // Update total
    const totalElement = document.getElementById('totalWorkHours');
    if (totalElement) {
        totalElement.textContent = totalHours.toFixed(2);
    }
    
    // Disable inputs (read-only for accounting)
    disableInputs();
}

/**
 * Format time from microseconds
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
 * Clear the time table
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
    
    const totalElement = document.getElementById('totalWorkHours');
    if (totalElement) {
        totalElement.textContent = '-';
    }
}

/**
 * Disable all inputs (read-only mode)
 */
function disableInputs() {
    const inputs = document.querySelectorAll('#timeTable input');
    inputs.forEach(input => {
        input.disabled = true;
    });
}

/**
 * Setup event listeners
 */
function setupEventListeners() {
    // Student dropdown
    const studentSelect = document.getElementById('studentSelect');
    if (studentSelect) {
        studentSelect.addEventListener('change', (e) => {
            onStudentChange(e.target.value);
        });
    }
    
    // Week navigation
    const prevBtn = document.getElementById('prevBtn');
    const nextBtn = document.getElementById('nextBtn');
    
    if (prevBtn) {
        prevBtn.addEventListener('click', () => navigateWeek(-1));
    }
    if (nextBtn) {
        nextBtn.addEventListener('click', () => navigateWeek(1));
    }
    
    // Submit button (Erledigt)
    const submitBtn = document.getElementById('submitBtn');
    if (submitBtn) {
        submitBtn.addEventListener('click', markAsProcessed);
    }
}

/**
 * Navigate to previous or next week
 */
async function navigateWeek(direction) {
    if (!currentStudent) return;
    
    const submissions = accountingService.studentSubmissions;
    const currentIndex = submissions.findIndex(s => s.id === currentSubmission?.id);
    
    const newIndex = currentIndex + direction;
    if (newIndex >= 0 && newIndex < submissions.length) {
        currentSubmission = submissions[newIndex];
        currentWeek = currentSubmission.week_number;
        currentYear = currentSubmission.year;
        
        updateWeekDisplay();
        updateTableDates();
        updateStatusDisplay();
        await loadTimeEntriesForSubmission();
        
        // Only enable submit button if status is 'bestaetigt'
        if (currentSubmission.status === 'bestaetigt') {
            enableSubmitButton();
        } else {
            disableSubmitButton();
        }
    }
}

/**
 * Mark current submission as processed (erledigt)
 */
async function markAsProcessed() {
    if (!currentSubmission) {
        showMessage('Kein Eintrag ausgewählt.', 'error');
        return;
    }
    
    try {
        await accountingService.markAsProcessed(currentSubmission.id);
        showMessage('Eintrag als erledigt markiert!', 'success');
        
        // Refresh data
        await accountingService.getApprovedSubmissions();
        
        // Update UI
        currentSubmission.status = 'erledigt';
        updateStatusDisplay();
        disableSubmitButton();
        
        // Move to next submission if available
        const submissions = accountingService.getSubmissionsForStudent(currentStudent.id);
        if (submissions.length > 0) {
            await onStudentChange(currentStudent.id);
        }
    } catch (error) {
        console.error('Error marking as processed:', error);
        showMessage('Fehler beim Markieren als erledigt.', 'error');
    }
}

/**
 * Show message to user
 */
function showMessage(message, type = 'info') {
    // Create or get message container
    let msgContainer = document.getElementById('messageContainer');
    if (!msgContainer) {
        msgContainer = document.createElement('div');
        msgContainer.id = 'messageContainer';
        msgContainer.style.cssText = 'position: fixed; top: 20px; right: 20px; z-index: 1000;';
        document.body.appendChild(msgContainer);
    }
    
    const msgDiv = document.createElement('div');
    msgDiv.className = `message message-${type}`;
    msgDiv.style.cssText = `
        padding: 12px 20px;
        margin-bottom: 10px;
        border-radius: 8px;
        color: white;
        font-weight: 500;
        box-shadow: 0 2px 10px rgba(0,0,0,0.2);
        animation: slideIn 0.3s ease;
    `;
    
    // Set background color based on type
    const colors = {
        success: '#4CAF50',
        error: '#f44336',
        info: '#2196F3',
        warning: '#ff9800'
    };
    msgDiv.style.backgroundColor = colors[type] || colors.info;
    msgDiv.textContent = message;
    
    msgContainer.appendChild(msgDiv);
    
    // Auto remove after 3 seconds
    setTimeout(() => {
        msgDiv.remove();
    }, 3000);
}

/**
 * Disable submit button
 */
function disableSubmitButton() {
    const submitBtn = document.getElementById('submitBtn');
    if (submitBtn) {
        submitBtn.disabled = true;
        submitBtn.style.opacity = '0.5';
    }
}

/**
 * Enable submit button
 */
function enableSubmitButton() {
    const submitBtn = document.getElementById('submitBtn');
    if (submitBtn) {
        submitBtn.disabled = false;
        submitBtn.style.opacity = '1';
    }
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
 * Get week date range helper
 */
function getWeekDateRange(weekNumber, year) {
    // Calculate the first day of the week
    const jan4 = new Date(year, 0, 4);
    const dayOfWeek = jan4.getDay() || 7;
    const firstMonday = new Date(jan4);
    firstMonday.setDate(jan4.getDate() - dayOfWeek + 1);
    
    const weekStart = new Date(firstMonday);
    weekStart.setDate(firstMonday.getDate() + (weekNumber - 1) * 7);
    
    const weekEnd = new Date(weekStart);
    weekEnd.setDate(weekStart.getDate() + 6);
    
    const formatDate = (date) => {
        const day = String(date.getDate()).padStart(2, '0');
        const month = String(date.getMonth() + 1).padStart(2, '0');
        return `${day}.${month}`;
    };
    
    return {
        start: formatDate(weekStart),
        end: formatDate(weekEnd)
    };
}
