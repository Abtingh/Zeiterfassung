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
    
    console.log(`${i + 1}. Woche: ${formatDate(monday)} - ${formatDate(sunday)} (Friday: ${formatDate(date)})`);
    
    const card = document.createElement('div');
    card.className = 'weekCard';
    
    // Add click functionality to navigate to time entry page
    card.addEventListener('click', function() {
      const weekNumber = i + 1;
      const month = month0 + 1; // Convert to 1-indexed
      const url = `student_ZeitEintragen.html?year=${year}&month=${month}&week=${weekNumber}`;
      window.location.href = url;
    });
    
    // Add hover effect
    card.style.cursor = 'pointer';

    const left = document.createElement('div');
    left.className = 'weekHolder';
    left.innerHTML = `<img src="../assets/iconWoche.svg" alt="">
    <p>${i + 1}. Woche</p>`;

    const right = document.createElement('div');
    // Example status assignment
    switch (i) {
      case 0:
        status = 'accepted';
        break;
      case 1:
        status = 'needReview';
        break;
      case 2:
        status = 'sent';
        break;
      default:
        status = 'open';
        break;
    }
    right.className = 'statusHolder';

    switch (status) {
      case 'accepted':
        right.classList.add('accepted');
        right.innerHTML = `
          <p>Bestätigt</p>
          <img src="../assets/iconAccepted.svg" alt="">
        `;
        break;

        case 'needReview':
        right.classList.add('needReview');
        right.innerHTML = `
          <p>Korrektur</p>
          <img src="../assets/iconNeedReview.svg" alt="">
        `;
        break;

        case 'sent':
        right.classList.add('sent');
        right.innerHTML = `
          <p>Gesendet</p>
          <img src="../assets/iconSent.svg" alt="">
        `;
        break;

        case 'open':
        right.classList.add('open');
        right.innerHTML = `
          <p>Offen</p>
          <img src="../assets/iconOpen.svg" alt="">
        `;
        break;

    default:
      right.innerHTML = `<p>—</p>`;
      break;
    }

    card.appendChild(left);
    card.appendChild(right);
    container.appendChild(card);
  });

  console.log(`Month ${month0 + 1}/${year} → ${fridays.length} weeks`);
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
