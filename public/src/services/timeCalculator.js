// Time calculation + week navigation
// NAV uses Friday-based week blocks (unchanged).
// LABEL uses Monday-based rule: in months with 5 Mondays, the first Monday is week 0.

document.addEventListener('DOMContentLoaded', function () {
  const timeTable    = document.getElementById('timeTable');
  const prevBtn      = document.getElementById('prevBtn');
  const nextBtn      = document.getElementById('nextBtn');
  const dateElement  = document.getElementById('date');
  const weekElement  = document.querySelector('#weekHolder p');

  if (!timeTable) return;

  // ---------- Helpers for Fridays (navigation stays the same) ----------
  /**
   * Generates all work week Fridays for a given month.
   * 
   * BUSINESS RULE: 
   * - Each month ends at the FIRST FRIDAY of the NEXT month
   * - Work weeks start on the Monday AFTER the first Friday of each month
   * 
   * Examples:
   * - September 2025 ends at Oct 3 (first Friday) → October starts Monday Oct 6
   * - October 2025 ends at Nov 7 (first Friday) → November starts Monday Nov 10
   * - November 2025 ends at Dec 5 (first Friday) → December starts Monday Dec 8
   * 
   * @param {number} year - The year (e.g., 2025)
   * @param {number} month0 - The month (0-11, where 0=January)
   * @returns {Date[]} Array of Friday dates representing work weeks for this work month
   */
  function fridaysInMonth(year, month0) {
    // Step 1: Find the first Friday of the month
    const daysInMonth = new Date(year, month0 + 1, 0).getDate(); // Total days in the month
    const firstDow = new Date(year, month0, 1).getDay(); // Day of week for 1st of month (0=Sunday, 5=Friday)
    const firstFridayDate = 1 + ((5 - firstDow + 7) % 7); // Calculate date of first Friday
    
    // If there's no Friday in this month (edge case), return empty array
    if (firstFridayDate > daysInMonth) {
      return []; // No Fridays in this month
    }
    
    const firstFriday = new Date(year, month0, firstFridayDate);
    
    // Step 2: Calculate work month start = Monday AFTER first Friday
    const workStartMonday = new Date(firstFriday);
    workStartMonday.setDate(firstFriday.getDate() + 3); // Friday + 3 days = next Monday
    
    // Step 3: Find next month's first Friday (to know when to stop)
    const nextMonth = month0 + 1 > 11 ? 0 : month0 + 1; // Wrap to January if December
    const nextYear = month0 + 1 > 11 ? year + 1 : year; // Increment year if wrapping
    const nextMonthDays = new Date(nextYear, nextMonth + 1, 0).getDate();
    const nextFirstDow = new Date(nextYear, nextMonth, 1).getDay();
    const nextFirstFridayDate = 1 + ((5 - nextFirstDow + 7) % 7);
    
    let nextFirstFriday;
    if (nextFirstFridayDate <= nextMonthDays) {
      nextFirstFriday = new Date(nextYear, nextMonth, nextFirstFridayDate);
    } else {
      // If no Friday in next month, extend to end of next month
      nextFirstFriday = new Date(nextYear, nextMonth + 1, 0);
    }
    
    // Step 4: Generate all Fridays from work start until next month's first Friday (inclusive)
    const workFridays = [];
    let currentFriday = new Date(workStartMonday);
    currentFriday.setDate(workStartMonday.getDate() + 4); // Monday + 4 days = Friday of that week
    
    // Loop through weeks, adding each Friday
    while (currentFriday <= nextFirstFriday) {
      workFridays.push(new Date(currentFriday)); // Store a copy of the date
      currentFriday.setDate(currentFriday.getDate() + 7); // Move to next Friday (+7 days)
    }
    
    return workFridays;
  }

  /**
   * Given a Friday date, returns an array of all 7 days in that week (Monday-Sunday).
   * 
   * @param {Date} fridayDate - The Friday of the week
   * @returns {Date[]} Array of 7 dates [Monday, Tuesday, ..., Sunday]
   */
  function getWeekDates(fridayDate) {
    const friday = new Date(fridayDate);
    const monday = new Date(friday);
    monday.setDate(friday.getDate() - 4); // Friday - 4 days = Monday
    
    const dates = [];
    for (let i = 0; i < 7; i++) {
      const dt = new Date(monday);
      dt.setDate(monday.getDate() + i); // Add i days to Monday
      dates.push(dt);
    }
    return dates; // Array: [Monday, Tuesday, Wednesday, Thursday, Friday, Saturday, Sunday]
  }

  // ---------- Helpers for Monday-based display number ----------
  /**
   * Finds all Mondays within a specific month.
   * 
   * @param {number} year - The year
   * @param {number} month0 - The month (0-11)
   * @returns {Date[]} Array of all Monday dates in the month
   */
  function mondaysInMonth(year, month0) {
    const daysInMonth = new Date(year, month0 + 1, 0).getDate(); // Get total days in month
    const out = [];
    
    // Loop through all days in the month
    for (let d = 1; d <= daysInMonth; d++) {
      const dt = new Date(year, month0, d);
      if (dt.getDay() === 1) out.push(dt); // 1 = Monday (0=Sunday, 1=Monday, ...)
    }
    return out;
  }

  /**
   * Checks if two dates are the same day (ignoring time).
   * 
   * @param {Date} a - First date
   * @param {Date} b - Second date
   * @returns {boolean} True if same year, month, and date
   */
  function sameDay(a, b) {
    return a.getFullYear() === b.getFullYear() &&
           a.getMonth() === b.getMonth() &&
           a.getDate() === b.getDate();
  }

  /**
   * Calculates the display week number based on work month rules:
   * Week numbering ALWAYS starts from 1 (never 0 for users).
   * The first work week (Monday after first Friday) is Week 1.
   * 
   * IMPORTANT: Uses the work month from currentWeek, not the Monday's calendar month!
   * This ensures that October Week 5 (Mon Nov 3) shows as "5" not "1".
   * 
   * @param {Date} mondayDate - The Monday of the week
   * @param {number} workYear - The work month's year
   * @param {number} workMonth - The work month (0-11)
   * @returns {number} The week number to display (always starts from 1)
   */
  function getDisplayWeekNumber(mondayDate, workYear, workMonth) {
    // Get all work week Fridays for the WORK month (not Monday's calendar month)
    const frs = fridaysInMonth(workYear, workMonth);
    
    // Find which work week this Monday belongs to
    for (let i = 0; i < frs.length; i++) {
      const friday = frs[i];
      const weekMonday = new Date(friday);
      weekMonday.setDate(friday.getDate() - 4); // Friday - 4 = Monday
      
      if (sameDay(weekMonday, mondayDate)) {
        return i + 1; // Return 1-based week number (1, 2, 3, 4, ...)
      }
    }
    
    return 1; // Default to week 1 if not found
  }

  // ---------- Format ----------
  /**
   * Formats a date as DD.MM (e.g., "01.11" for November 1st).
   * 
   * @param {Date} date - The date to format
   * @returns {string} Formatted date string "DD.MM"
   */
  function formatDate(date) {
    const d = String(date.getDate()).padStart(2, '0'); // Day with leading zero
    const m = String(date.getMonth() + 1).padStart(2, '0'); // Month with leading zero (month is 0-indexed)
    return `${d}.${m}`;
  }

  // ---------- Current week state ----------
  /**
   * Determines which work week contains today's date.
   * Searches first in the current month, then falls back to previous month
   * for edge cases (e.g., if today is Nov 1 but the work week started Oct 27).
   * 
   * @returns {Object} { year, month (0-11), weekIndex (0-based), fridayDate }
   */
  function getCurrentWeekInfo() {
    const today = new Date();
    today.setHours(0, 0, 0, 0); // Normalize to midnight (start of day)
    const year = today.getFullYear();
    const month = today.getMonth();
    
    // STEP 1: Check current month's work weeks
    let frs = fridaysInMonth(year, month);
    for (let i = 0; i < frs.length; i++) {
      const friday = frs[i];
      
      // Calculate Monday and Sunday of this work week
      const monday = new Date(friday);
      monday.setDate(friday.getDate() - 4); // Friday - 4 = Monday
      monday.setHours(0, 0, 0, 0); // Start of Monday
      
      const sunday = new Date(friday);
      sunday.setDate(friday.getDate() + 2); // Friday + 2 = Sunday
      sunday.setHours(23, 59, 59, 999); // End of Sunday
      
      // Check if today falls within this week (Monday to Sunday inclusive)
      if (today >= monday && today <= sunday) {
        return { year, month, weekIndex: i, fridayDate: null };
      }
    }
    
    // STEP 2: If not found, check previous month (edge case at start of month)
    // Example: Nov 1 (Saturday) might belong to a week that started Oct 27 (Monday)
    const prevMonth = month === 0 ? 11 : month - 1; // Wrap to December if January
    const prevYear = month === 0 ? year - 1 : year; // Decrement year if wrapping
    frs = fridaysInMonth(prevYear, prevMonth);
    
    for (let i = 0; i < frs.length; i++) {
      const friday = frs[i];
      
      const monday = new Date(friday);
      monday.setDate(friday.getDate() - 4);
      monday.setHours(0, 0, 0, 0);
      
      const sunday = new Date(friday);
      sunday.setDate(friday.getDate() + 2);
      sunday.setHours(23, 59, 59, 999);
      
      // Check if today falls in this previous month's week
      if (today >= monday && today <= sunday) {
        return { year: prevYear, month: prevMonth, weekIndex: i, fridayDate: null };
      }
    }
    
    // STEP 3: Default fallback - return first week of current month
    // (This should rarely happen unless there's a data issue)
    return { year, month, weekIndex: 0, fridayDate: null };
  }

  // Initialize with the current week containing today's date
  let currentWeek = getCurrentWeekInfo();

  // ---------- UI update ----------
  /**
   * Updates the week display (date range and week number) in the UI.
   * Handles boundary conditions (moving to next/previous month).
   */
  function updateWeekDisplay() {
    const frs = fridaysInMonth(currentWeek.year, currentWeek.month);

    // BOUNDARY CHECK 1: If weekIndex is beyond the last week of the month
    if (currentWeek.weekIndex >= frs.length) {
      currentWeek.month++; // Move to next month
      if (currentWeek.month > 11) { 
        currentWeek.month = 0; // Wrap to January
        currentWeek.year++; // Increment year
      }
      currentWeek.weekIndex = 0; // Start at first week of new month
      return updateWeekDisplay(); // Recursively call to update display
    }
    
    // BOUNDARY CHECK 2: If weekIndex is negative (navigated before first week)
    if (currentWeek.weekIndex < 0) {
      currentWeek.month--; // Move to previous month
      if (currentWeek.month < 0) { 
        currentWeek.month = 11; // Wrap to December
        currentWeek.year--; // Decrement year
      }
      const prevFrs = fridaysInMonth(currentWeek.year, currentWeek.month);
      currentWeek.weekIndex = prevFrs.length - 1; // Jump to last week of previous month
      return updateWeekDisplay(); // Recursively call to update display
    }

    // Get the Friday of the current week
    currentWeek.fridayDate = frs[currentWeek.weekIndex];
    
    // Get all 7 days of this week (Monday-Sunday)
    const weekDates = getWeekDates(currentWeek.fridayDate);

    const startDate = weekDates[0]; // Monday
    const endDate   = weekDates[6]; // Sunday

    // UPDATE HEADER: Display the date range (e.g., "27.10 - 02.11")
    dateElement.textContent = ` ${formatDate(startDate)} - ${formatDate(endDate)}`;
    
    // UPDATE WEEK NUMBER: Calculate and display the week number (e.g., "2.Woche")
    // Pass the work month info so it shows correct week number (e.g., Oct Week 5, not Nov Week 1)
    const displayWeekNo = getDisplayWeekNumber(startDate, currentWeek.year, currentWeek.month);
    weekElement.textContent = `${displayWeekNo}.Woche`;

    // UPDATE TABLE: Fill in the date for each day column in the table
    updateTableDates(weekDates);
  }

  /**
   * Updates the date displayed in each day column header of the time table.
   * 
   * @param {Date[]} weekDates - Array of 7 dates (Monday-Sunday)
   */
  function updateTableDates(weekDates) {
    // Select all 7 day column headers
    const dayEls = [
      document.querySelector('.first.day'),
      document.querySelector('.second.day'),
      document.querySelector('.third.day'),
      document.querySelector('.fourth.day'),
      document.querySelector('.fifth.day'),
      document.querySelector('.sixth.day'),
      document.querySelector('.seventh.day')
    ];
    
    // Update each column with the corresponding date
    dayEls.forEach((el, i) => { 
      if (el && weekDates[i]) {
        el.textContent = formatDate(weekDates[i]); // Format as DD.MM
      }
    });
  }

  /**
   * Navigate to the previous or next week.
   * 
   * @param {number} direction - -1 for previous week, +1 for next week
   */
  function navigateWeek(direction) {
    currentWeek.weekIndex += direction; // Increment or decrement week index
    updateWeekDisplay(); // Update the UI
    resetTable(); // Clear all input fields
  }

  /**
   * Initialize the week display from URL parameters (if provided).
   * Example: ?year=2025&month=11&week=2
   * This allows deep linking to a specific week.
   */
  function initializeFromURL() {
    const url = new URLSearchParams(window.location.search);
    const year = url.get('year');
    const month = url.get('month');
    const week = url.get('week');
    
    if (year && month && week) {
      currentWeek.year = parseInt(year, 10); // Parse year as integer
      currentWeek.month = parseInt(month, 10) - 1; // Convert to 0-indexed (URL uses 1-12)
      currentWeek.weekIndex = parseInt(week, 10) - 1; // Convert to 0-indexed (URL uses 1-based)
    }
  }

  // EVENT LISTENERS: Wire up the previous/next week buttons
  if (prevBtn) prevBtn.addEventListener('click', () => navigateWeek(-1)); // Go to previous week
  if (nextBtn) nextBtn.addEventListener('click', () => navigateWeek(1)); // Go to next week

  // ---------- Time calc ----------
  // Select all enabled time and pause input fields
  const timeInputs  = timeTable.querySelectorAll('.time-input:not([disabled])');
  const pauseInputs = timeTable.querySelectorAll('.pause-input');

  // Add event listeners to recalculate hours when inputs change
  timeInputs.forEach(i => { 
    i.addEventListener('change', calculateRowHours); 
    i.addEventListener('blur', calculateRowHours); 
  });
  pauseInputs.forEach(i => { 
    i.addEventListener('change', calculateRowHours); 
    i.addEventListener('blur', calculateRowHours); 
  });

  /**
   * Calculates and displays the hours for a single row in the time table.
   * Triggered when a time or pause input changes.
   * 
   * @param {Event} event - The change/blur event from the input field
   */
  function calculateRowHours(event) {
    const row = event.target.closest('tr'); // Find the parent row
    if (!row) return;

    // Get all the input fields and display cells for this row
    const fromInput    = row.querySelector('.time-input:first-of-type'); // Start time (e.g., "08:00")
    const toInput      = row.querySelector('.time-input:last-of-type');  // End time (e.g., "17:00")
    const pauseInput   = row.querySelector('.pause-input');              // Pause hours (e.g., "1.0")
    const hoursCell    = row.querySelector('.calculated-hours');         // Total hours display
    const workHoursCell= row.querySelector('.work-hours');               // Work hours display (total - pause)

    if (!fromInput || !toInput || !pauseInput || !hoursCell || !workHoursCell) return;

    const fromTime   = fromInput.value;  // e.g., "08:00"
    const toTime     = toInput.value;    // e.g., "17:00"
    const pauseHours = parseFloat(pauseInput.value) || 0; // e.g., 1.0

    // If both start and end times are provided, calculate hours
    if (fromTime && toTime) {
      const totalHours   = calculateTimeDifference(fromTime, toTime); // e.g., 9.0 hours
      const workingHours = Math.max(0, totalHours - pauseHours);      // e.g., 8.0 hours (9 - 1)
      
      hoursCell.textContent = totalHours.toFixed(1);    // Display total hours
      workHoursCell.textContent = workingHours.toFixed(1); // Display work hours
    } else {
      // If times are incomplete, show dashes
      hoursCell.textContent = '-';
      workHoursCell.textContent = '-';
    }
    
    // Recalculate the total for all rows
    calculateTotalHours();
  }

  /**
   * Calculates the time difference in hours between two time strings.
   * Handles overnight shifts (e.g., 22:00 to 06:00).
   * 
   * @param {string} startTime - Start time in HH:MM format (e.g., "08:00")
   * @param {string} endTime - End time in HH:MM format (e.g., "17:00")
   * @returns {number} Time difference in hours (e.g., 9.0)
   */
  function calculateTimeDifference(startTime, endTime) {
    // Create Date objects (date doesn't matter, we just care about time)
    const start = new Date(`2000-01-01T${startTime}`);
    const end   = new Date(`2000-01-01T${endTime}`);
    
    // Handle overnight shifts: if end is before start, add 1 day to end
    if (end < start) end.setDate(end.getDate() + 1);
    
    // Calculate difference in milliseconds, then convert to hours
    const diffMs = end - start;
    return Math.max(0, diffMs / (1000 * 60 * 60)); // Convert ms to hours
  }

  /**
   * Calculates and displays the total work hours across all rows.
   * Sums up all "work hours" values from the table.
   */
  function calculateTotalHours() {
    const workHoursCells = timeTable.querySelectorAll('.work-hours'); // All work hour cells
    const totalElement   = document.getElementById('totalWorkHours'); // Total display element
    if (!totalElement) return;
    
    let total = 0;
    // Sum up all work hours
    workHoursCells.forEach(cell => {
      const v = parseFloat(cell.textContent);
      if (!isNaN(v)) total += v;
    });
    
    // Display the total with 1 decimal place
    totalElement.textContent = total.toFixed(1);
  }

  // ---------- Reset ----------
  /**
   * Resets all input fields and calculated values in the time table.
   * Clears: time inputs, pause inputs, notes, calculated hours, and total.
   */
  function resetTable() {
    // Select all enabled input fields
    const enabledTimeInputs  = timeTable.querySelectorAll('.time-input:not([disabled])');
    const enabledPauseInputs = timeTable.querySelectorAll('.pause-input');
    const enabledNotesInputs = timeTable.querySelectorAll('.notes-input:not([disabled])');

    // Clear all input values
    enabledTimeInputs.forEach(i => (i.value = ''));
    enabledPauseInputs.forEach(i => (i.value = ''));
    enabledNotesInputs.forEach(i => (i.value = ''));

    // Reset calculated hours display for each row
    const rows = timeTable.querySelectorAll('tbody tr:not(.total-row)');
    rows.forEach(row => {
      const timeInput = row.querySelector('.time-input');
      if (timeInput && !timeInput.disabled) {
        const hoursCell = row.querySelector('.calculated-hours'); // Total hours cell
        const workCell  = row.querySelector('.work-hours');       // Work hours cell
        if (hoursCell) hoursCell.textContent = '-';
        if (workCell)  workCell.textContent = '-';
      }
    });

    // Reset the total work hours display
    const totalElement = document.getElementById('totalWorkHours');
    if (totalElement) totalElement.textContent = '-';
  }

  // EVENT LISTENER: Wire up the reset button
  const resetBtn = document.getElementById('resetBtn');
  if (resetBtn) resetBtn.addEventListener('click', () => resetTable());

  // ---------- Submit (demo) ----------
  const submitBtn = document.getElementById('submitBtn');
  if (submitBtn) {
    submitBtn.addEventListener('click', function () {
      const tableData = collectTableData(); // Collect all data from the table
      console.log('Submitting time data:', tableData); // Log to console for debugging
      alert('Zeiterfassung wurde erfolgreich gesendet!'); // Show success message
    });
  }

  /**
   * Collects all time tracking data from the table.
   * Returns a structured object with all week info and daily entries.
   * 
   * @returns {Object} Complete time tracking data for the current week
   */
  function collectTableData() {
    const rows = timeTable.querySelectorAll('tbody tr:not(.total-row)'); // Get all data rows
    const data = [];
    
    // Extract data from each row
    rows.forEach(row => {
      const day       = row.querySelector('.day-label').textContent; // e.g., "Montag"
      const date      = row.querySelector('.day').textContent;       // e.g., "27.10"
      const fromTime  = row.querySelector('.time-input:first-of-type').value; // e.g., "08:00"
      const toTime    = row.querySelector('.time-input:last-of-type').value;  // e.g., "17:00"
      const pauseIn   = row.querySelector('.pause-input');
      const pause     = pauseIn ? pauseIn.value : '0'; // Pause hours
      const notes     = row.querySelector('.notes-input').value; // Notes/comments
      const workHours = row.querySelector('.work-hours').textContent; // Calculated work hours

      data.push({
        day,
        date,
        from: fromTime,
        to: toTime,
        pause: parseFloat(pause) || 0,
        notes,
        workHours: workHours !== '-' ? parseFloat(workHours) : 0
      });
    });

    // Get the week information for the export
    const weekDates = getWeekDates(fridaysInMonth(currentWeek.year, currentWeek.month)[currentWeek.weekIndex]);
    const displayWeekNo = getDisplayWeekNumber(weekDates[0], currentWeek.year, currentWeek.month); // Week number for display

    // Return complete dataset
    return {
      year: currentWeek.year,              // e.g., 2025
      month: currentWeek.month + 1,        // e.g., 11 (month as 1-12)
      weekNumber: displayWeekNo,           // e.g., 2
      weekRange: dateElement.textContent.trim(), // e.g., "27.10 - 02.11"
      weekData: data,                      // Array of daily entries
      totalHours: document.getElementById('totalWorkHours').textContent // Total work hours
    };
  }

  // ---------- Init ----------
  // Initialize the page: check URL parameters and display the current week
  initializeFromURL();  // Load week from URL if parameters exist
  updateWeekDisplay();  // Render the week display
});
