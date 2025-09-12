
function fridaysInMonth(year, month0) {
  const daysInMonth = new Date(year, month0 + 1, 0).getDate();
  const firstDow = new Date(year, month0, 1).getDay(); 
  const firstFridayDate = 1 + ((5 - firstDow + 7) % 7); 

  const fridays = [];
  for (let d = firstFridayDate; d <= daysInMonth; d += 7) {
    fridays.push(new Date(year, month0, d));
  }
  return fridays; 
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

  fridays.forEach((date, i) => {
    const card = document.createElement('div');
    card.className = 'weekCard';

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
