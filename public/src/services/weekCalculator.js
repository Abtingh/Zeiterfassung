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
 * Calculates the display week number based on work month rules.
 * Week numbering ALWAYS starts from 1 (never 0 for users).
 * 
 * @param {Date} mondayDate - The Monday of the week
 * @param {number} workYear - The work month's year
 * @param {number} workMonth - The work month (0-11)
 * @returns {number} The week number to display (always starts from 1)
 */
function getDisplayWeekNumber(mondayDate, workYear, workMonth) {
  // Get all work week Fridays for the WORK month
  const frs = fridaysInMonth(workYear, workMonth);
  
  // Find which work week this Monday belongs to
  for (let i = 0; i < frs.length; i++) {
    const friday = frs[i];
    const weekMonday = new Date(friday);
    weekMonday.setDate(friday.getDate() - 4); // Friday - 4 = Monday
    
    if (sameDay(weekMonday, mondayDate)) {
      return i + 1; // Week numbers start from 1
    }
  }
  
  return 1; // Default to week 1 if not found
}

/**
 * Checks if two dates are the same day (ignoring time).
 */
function sameDay(a, b) {
  return a.getFullYear() === b.getFullYear() &&
         a.getMonth() === b.getMonth() &&
         a.getDate() === b.getDate();
}


/**
 * Renders a list of week cards for the specified month and year.
 * Each card displays the week number and a status placeholder.
 * The weeks are based on Fridays found in the given month.
 * 
 * @param {number} year - The year (e.g., 2023)
 * @param {number} month0 - The month as zero-based index (0 = January, 11 = December)
 * @returns {void}
 * 
 * @example
 * // Render weeks for March 2023
 * renderMonthWeeks(2023, 2);
 * 
 * @throws {Error} Logs error to console if 'weeksList' element is not found in DOM
 */

function renderMonthWeeks(year, month0) {
  const container = document.getElementById('wochenList');
  if (!container) {
    console.error("Element with id 'wochenList' not found in the DOM.");
    return;
  }
  container.innerHTML = '';

  const fridays = fridaysInMonth(year, month0);
  
  // Helper function to get week dates based on Friday
  function getWeekDates(fridayDate) {
    const friday = new Date(fridayDate);
    const monday = new Date(friday);
    monday.setDate(friday.getDate() - 4); // Friday - 4 days = Monday
    
    const dates = [];
    for (let i = 0; i < 7; i++) {
      const date = new Date(monday);
      date.setDate(monday.getDate() + i);
      dates.push(date);
    }
    return dates;
  }
  
  // Helper function to format date as DD.MM
  function formatDate(date) {
    const day = date.getDate().toString().padStart(2, '0');
    const month = (date.getMonth() + 1).toString().padStart(2, '0');
    return `${day}.${month}`;
  }
  
  console.log(`=== ${new Intl.DateTimeFormat('de-DE', { month: 'long', year: 'numeric' }).format(new Date(year, month0))} Week Ranges ===`);

  fridays.forEach((date, i) => {
    const weekDates = getWeekDates(date);
    const monday = weekDates[0]; // Monday
    const sunday = weekDates[6]; // Sunday
    
    // Calculate the proper display week number using the same logic as timeCalculator
    const displayWeekNo = getDisplayWeekNumber(monday, year, month0);
    
    console.log(`${displayWeekNo}. Woche: ${formatDate(monday)} - ${formatDate(sunday)} (Friday: ${formatDate(date)})`);
    
    const card = document.createElement('div');
    card.className = 'weekCard';
    
    // Add click functionality to navigate to time entry page with correct week index
    card.addEventListener('click', function() {
      const month = month0 + 1; // Convert to 1-indexed
      const weekIndex = i + 1; // Week index within the month (1-based)
      const url = `/ZeitEintragen?year=${year}&month=${month}&week=${weekIndex}`;
      console.log(`Navigating to: ${url}`);
      window.location.href = url;
    });
    
    // Add hover effect
    card.style.cursor = 'pointer';

    const left = document.createElement('div');
    left.className = 'weekHolder';
    left.innerHTML = `<img src="../assets/iconWoche.svg" alt="">
    <p>${displayWeekNo}. Woche</p>`;

    const right = document.createElement('div');
    right.className = 'statusHolder';
    
    // Fetch status from backend for this specific week
    // For now, set to loading state
    right.innerHTML = `<p>Laden...</p>`;
    
    // Calculate globally unique week number: (month * 5) + weekIndex
    // This ensures October Week 4 != November Week 4
    const globalWeekNumber = (month0 * 5) + displayWeekNo;
    console.log(`Month ${month0}, Week ${displayWeekNo} → Global Week ${globalWeekNumber}`);
    
    // Async fetch the status with the global week number
    fetchWeekStatus(globalWeekNumber, year).then(status => {
      updateStatusDisplay(right, status);
    }).catch(() => {
      // Default to 'open' if fetch fails
      updateStatusDisplay(right, 'offen');
    });

    card.appendChild(left);
    card.appendChild(right);
    container.appendChild(card);
  });

  console.log(`Month ${month0 + 1}/${year} → ${fridays.length} weeks`);
}

/**
 * Calculates a globally unique week number within the year
 * Based on month and week within that month
 * Formula: (month * 10) + weekInMonth
 * Examples: October Week 4 = 94, November Week 1 = 101
 */
function calculateGlobalWeekNumber(month0, weekInMonth) {
  return ((month0 + 1) * 10) + weekInMonth;
}

/**
 * Fetches the status of a specific week from the backend
 */
async function fetchWeekStatus(weekNumber, year) {
  try {
    const response = await fetch(`/api/time-entries/week?week=${weekNumber}&year=${year}`, {
      method: 'GET',
      credentials: 'include'
    });

    if (!response.ok) {
      return 'offen'; // Default to open if request fails
    }

    const data = await response.json();
    return data.status || 'offen';
  } catch (error) {
    console.error('Error fetching week status:', error);
    return 'offen';
  }
}

/**
 * Updates the status display with the proper styling and icon
 */
function updateStatusDisplay(element, status) {
  // Clear existing classes
  element.className = 'statusHolder';
  
  switch (status) {
    case 'bestaetigt':
      element.classList.add('accepted');
      element.innerHTML = `
        <p>Bestätigt</p>
        <img src="/assets/iconAccepted.svg" alt="">
      `;
      break;

    case 'gesendet':
      element.classList.add('sent');
      element.innerHTML = `
        <p>Gesendet</p>
        <img src="/assets/iconSent.svg" alt="">
      `;
      break;

    case 'offen':
      element.classList.add('open');
      element.innerHTML = `
        <p>Offen</p>
        <img src="/assets/iconOpen.svg" alt="">
      `;
      break;

    case 'erledigt':
      element.classList.add('needReview');
      element.innerHTML = `
        <p>Erledigt</p>
        <img src="/assets/iconNeedReview.svg" alt="">
      `;
      break;

    default:
      element.innerHTML = `<p>—</p>`;
      break;
  }
}

const monthLabel = document.getElementById('monthLabel');
let currentDate = new Date(); 

const deMonth = new Intl.DateTimeFormat('de-DE', {
  month: 'long',
  year: 'numeric'
});

function renderMonth() {
  // نمایش ماه
  monthLabel.textContent = deMonth.format(currentDate);

  // تعداد هفته‌ها را حساب کن و لیست بساز
  renderMonthWeeks(currentDate.getFullYear(), currentDate.getMonth());
}

document.getElementById('prevBtn').addEventListener('click', () => {
  currentDate.setMonth(currentDate.getMonth() - 1);
  renderMonth();
});

document.getElementById('nextBtn').addEventListener('click', () => {
  currentDate.setMonth(currentDate.getMonth() + 1);
  renderMonth();
});

document.addEventListener('DOMContentLoaded', () => {
  renderMonth(); // اولین بار ماه جاری
});
