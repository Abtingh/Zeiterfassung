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
  function fridaysInMonth(year, month0) {
    // Find the first Friday of the month
    const daysInMonth = new Date(year, month0 + 1, 0).getDate();
    const firstDow = new Date(year, month0, 1).getDay(); 
    const firstFridayDate = 1 + ((5 - firstDow + 7) % 7);
    
    if (firstFridayDate > daysInMonth) {
      return []; // No Fridays in this month
    }
    
    const firstFriday = new Date(year, month0, firstFridayDate);
    
    // Work month starts on Monday after first Friday
    const workStartMonday = new Date(firstFriday);
    workStartMonday.setDate(firstFriday.getDate() + 3); // Friday + 3 days = Monday
    
    // Find next month's first Friday
    const nextMonth = month0 + 1 > 11 ? 0 : month0 + 1;
    const nextYear = month0 + 1 > 11 ? year + 1 : year;
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
    
    // Generate all Fridays from work start until and INCLUDING the week with next month's first Friday
    const workFridays = [];
    let currentFriday = new Date(workStartMonday);
    currentFriday.setDate(workStartMonday.getDate() + 4); // Monday + 4 = Friday
    
    while (currentFriday <= nextFirstFriday) {
      workFridays.push(new Date(currentFriday));
      currentFriday.setDate(currentFriday.getDate() + 7); // Next Friday
    }
    
    return workFridays;
  }

  function getWeekDates(fridayDate) {
    const friday = new Date(fridayDate);
    const monday = new Date(friday);
    monday.setDate(friday.getDate() - 4);
    const dates = [];
    for (let i = 0; i < 7; i++) {
      const dt = new Date(monday);
      dt.setDate(monday.getDate() + i);
      dates.push(dt);
    }
    return dates; // Monday..Sunday
  }

  // ---------- Helpers for Monday-based display number ----------
  function mondaysInMonth(year, month0) {
    const daysInMonth = new Date(year, month0 + 1, 0).getDate();
    const out = [];
    for (let d = 1; d <= daysInMonth; d++) {
      const dt = new Date(year, month0, d);
      if (dt.getDay() === 1) out.push(dt); // 1 = Monday
    }
    return out;
  }

  function sameDay(a, b) {
    return a.getFullYear() === b.getFullYear() &&
           a.getMonth() === b.getMonth() &&
           a.getDate() === b.getDate();
  }

  // If the month has 5 Mondays → first Monday labeled 0, second Monday labeled 1, etc.
  // If the month has 4 Mondays → first Monday labeled 1, second 2, etc.
  function getDisplayWeekNumber(mondayDate) {
    const y = mondayDate.getFullYear();
    const m = mondayDate.getMonth();
    const ms = mondaysInMonth(y, m);
    const idx = ms.findIndex(d => sameDay(d, mondayDate));
    if (idx < 0) return 1;
    const hasFive = ms.length === 5;
    return hasFive ? idx : idx + 1;
  }

  // ---------- Format ----------
  function formatDate(date) {
    const d = String(date.getDate()).padStart(2, '0');
    const m = String(date.getMonth() + 1).padStart(2, '0');
    return `${d}.${m}`;
  }

  // ---------- Current week state ----------
  function getCurrentWeekInfo() {
    const today = new Date();
    const year = today.getFullYear();
    const month = today.getMonth();
    const frs = fridaysInMonth(year, month);

    let currentWeekIndex = 0;
    for (let i = 0; i < frs.length; i++) {
      const friday = frs[i];
      const monday = new Date(friday);
      monday.setDate(friday.getDate() - 4);
      const sunday = new Date(friday);
      sunday.setDate(friday.getDate() + 2);
      if (today >= monday && today <= sunday) {
        currentWeekIndex = i;
        break;
      }
    }
    return { year, month, weekIndex: currentWeekIndex, fridayDate: null };
  }

  let currentWeek = getCurrentWeekInfo();

  // ---------- UI update ----------
  function updateWeekDisplay() {
    const frs = fridaysInMonth(currentWeek.year, currentWeek.month);

    // bounds
    if (currentWeek.weekIndex >= frs.length) {
      currentWeek.month++;
      if (currentWeek.month > 11) { currentWeek.month = 0; currentWeek.year++; }
      currentWeek.weekIndex = 0;
      return updateWeekDisplay();
    }
    if (currentWeek.weekIndex < 0) {
      currentWeek.month--;
      if (currentWeek.month < 0) { currentWeek.month = 11; currentWeek.year--; }
      const prevFrs = fridaysInMonth(currentWeek.year, currentWeek.month);
      currentWeek.weekIndex = prevFrs.length - 1;
      return updateWeekDisplay();
    }

    currentWeek.fridayDate = frs[currentWeek.weekIndex];
    const weekDates = getWeekDates(currentWeek.fridayDate);

    const startDate = weekDates[0]; // Monday
    const endDate   = weekDates[6]; // Sunday

    // header
    dateElement.textContent = ` ${formatDate(startDate)} - ${formatDate(endDate)}`;
    const displayWeekNo = getDisplayWeekNumber(startDate);
    weekElement.textContent = `${displayWeekNo}.Woche`;

    // fill table heads
    updateTableDates(weekDates);
  }

  function updateTableDates(weekDates) {
    const dayEls = [
      document.querySelector('.first.day'),
      document.querySelector('.second.day'),
      document.querySelector('.third.day'),
      document.querySelector('.fourth.day'),
      document.querySelector('.fifth.day'),
      document.querySelector('.sixth.day'),
      document.querySelector('.seventh.day')
    ];
    dayEls.forEach((el, i) => { if (el && weekDates[i]) el.textContent = formatDate(weekDates[i]); });
  }

  function navigateWeek(direction) {
    currentWeek.weekIndex += direction;
    updateWeekDisplay();
    resetTable();
  }

  function initializeFromURL() {
    const url = new URLSearchParams(window.location.search);
    const year = url.get('year');
    const month = url.get('month');
    const week = url.get('week');
    if (year && month && week) {
      currentWeek.year = parseInt(year, 10);
      currentWeek.month = parseInt(month, 10) - 1;
      currentWeek.weekIndex = parseInt(week, 10) - 1;
    }
  }

  if (prevBtn) prevBtn.addEventListener('click', () => navigateWeek(-1));
  if (nextBtn) nextBtn.addEventListener('click', () => navigateWeek(1));

  // ---------- Time calc ----------
  const timeInputs  = timeTable.querySelectorAll('.time-input:not([disabled])');
  const pauseInputs = timeTable.querySelectorAll('.pause-input');

  timeInputs.forEach(i => { i.addEventListener('change', calculateRowHours); i.addEventListener('blur', calculateRowHours); });
  pauseInputs.forEach(i => { i.addEventListener('change', calculateRowHours); i.addEventListener('blur', calculateRowHours); });

  function calculateRowHours(event) {
    const row = event.target.closest('tr');
    if (!row) return;

    const fromInput    = row.querySelector('.time-input:first-of-type');
    const toInput      = row.querySelector('.time-input:last-of-type');
    const pauseInput   = row.querySelector('.pause-input');
    const hoursCell    = row.querySelector('.calculated-hours');
    const workHoursCell= row.querySelector('.work-hours');

    if (!fromInput || !toInput || !pauseInput || !hoursCell || !workHoursCell) return;

    const fromTime   = fromInput.value;
    const toTime     = toInput.value;
    const pauseHours = parseFloat(pauseInput.value) || 0;

    if (fromTime && toTime) {
      const totalHours   = calculateTimeDifference(fromTime, toTime);
      const workingHours = Math.max(0, totalHours - pauseHours);
      hoursCell.textContent = totalHours.toFixed(1);
      workHoursCell.textContent = workingHours.toFixed(1);
    } else {
      hoursCell.textContent = '-';
      workHoursCell.textContent = '-';
    }
    calculateTotalHours();
  }

  function calculateTimeDifference(startTime, endTime) {
    const start = new Date(`2000-01-01T${startTime}`);
    const end   = new Date(`2000-01-01T${endTime}`);
    if (end < start) end.setDate(end.getDate() + 1); // overnight
    const diffMs = end - start;
    return Math.max(0, diffMs / (1000 * 60 * 60));
  }

  function calculateTotalHours() {
    const workHoursCells = timeTable.querySelectorAll('.work-hours');
    const totalElement   = document.getElementById('totalWorkHours');
    if (!totalElement) return;
    let total = 0;
    workHoursCells.forEach(cell => {
      const v = parseFloat(cell.textContent);
      if (!isNaN(v)) total += v;
    });
    totalElement.textContent = total.toFixed(1);
  }

  // ---------- Reset ----------
  function resetTable() {
    const enabledTimeInputs  = timeTable.querySelectorAll('.time-input:not([disabled])');
    const enabledPauseInputs = timeTable.querySelectorAll('.pause-input');
    const enabledNotesInputs = timeTable.querySelectorAll('.notes-input:not([disabled])');

    enabledTimeInputs.forEach(i => (i.value = ''));
    enabledPauseInputs.forEach(i => (i.value = ''));
    enabledNotesInputs.forEach(i => (i.value = ''));

    const rows = timeTable.querySelectorAll('tbody tr:not(.total-row)');
    rows.forEach(row => {
      const timeInput = row.querySelector('.time-input');
      if (timeInput && !timeInput.disabled) {
        const hoursCell = row.querySelector('.calculated-hours');
        const workCell  = row.querySelector('.work-hours');
        if (hoursCell) hoursCell.textContent = '-';
        if (workCell)  workCell.textContent = '-';
      }
    });

    const totalElement = document.getElementById('totalWorkHours');
    if (totalElement) totalElement.textContent = '-';
  }

  const resetBtn = document.getElementById('resetBtn');
  if (resetBtn) resetBtn.addEventListener('click', () => resetTable());

  // ---------- Submit (demo) ----------
  const submitBtn = document.getElementById('submitBtn');
  if (submitBtn) {
    submitBtn.addEventListener('click', function () {
      const tableData = collectTableData();
      console.log('Submitting time data:', tableData);
      alert('Zeiterfassung wurde erfolgreich gesendet!');
    });
  }

  function collectTableData() {
    const rows = timeTable.querySelectorAll('tbody tr:not(.total-row)');
    const data = [];
    rows.forEach(row => {
      const day       = row.querySelector('.day-label').textContent;
      const date      = row.querySelector('.day').textContent;
      const fromTime  = row.querySelector('.time-input:first-of-type').value;
      const toTime    = row.querySelector('.time-input:last-of-type').value;
      const pauseIn   = row.querySelector('.pause-input');
      const pause     = pauseIn ? pauseIn.value : '0';
      const notes     = row.querySelector('.notes-input').value;
      const workHours = row.querySelector('.work-hours').textContent;

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

    // Use the Monday of the current view for the exported label too
    const weekDates = getWeekDates(fridaysInMonth(currentWeek.year, currentWeek.month)[currentWeek.weekIndex]);
    const displayWeekNo = getDisplayWeekNumber(weekDates[0]);

    return {
      year: currentWeek.year,
      month: currentWeek.month + 1,
      weekNumber: displayWeekNo,
      weekRange: dateElement.textContent.trim(),
      weekData: data,
      totalHours: document.getElementById('totalWorkHours').textContent
    };
  }

  // ---------- Init ----------
  initializeFromURL();
  updateWeekDisplay();
});
