/**
 * Student Dropdown for Supervisor Home Page
 * Fetches students from server and displays their names
 */

let students = [];
const monthFmt = new Intl.DateTimeFormat('de-DE', { month: 'long', year: 'numeric' });
let viewDate = new Date();
let currentStudent = null;

/**
 * Fetch students from the server
 */
async function fetchStudents() {
    try {
        const response = await fetch('/api/supervisor/students', {
            method: 'GET',
            credentials: 'include'
        });

        if (!response.ok) {
            throw new Error('Failed to fetch students');
        }

        const data = await response.json();
        students = data || [];
        return students;
    } catch (error) {
        console.error('Error fetching students:', error);
        return [];
    }
}

/**
 * Get full name from student object
 */
function getStudentName(student) {
    const firstName = student.first_name?.String || student.first_name || '';
    const lastName = student.last_name?.String || student.last_name || '';
    return `${firstName} ${lastName}`.trim() || student.email;
}

/**
 * Populate the student dropdown
 */
function fillStudentSelect() {
    const sel = document.getElementById('studentSelect');
    if (!sel) return;

    if (students.length === 0) {
        sel.innerHTML = '<option value="" disabled selected>Keine Studenten verfügbar</option>';
        return;
    }

    sel.innerHTML = students.map(s => 
        `<option value="${s.id}">${getStudentName(s)}</option>`
    ).join('');

    // Select first student by default
    if (students.length > 0 && !currentStudent) {
        currentStudent = students[0].id;
    }
    sel.value = currentStudent;
}

function renderHeader() {
    const monthLabel = document.getElementById('monthLabel');
    if (monthLabel) {
        monthLabel.textContent = monthFmt.format(viewDate);
    }
}

function renderAll() {
    renderHeader();
    // If renderMonthWeeks exists (from weekCalculator.js), call it
    if (typeof renderMonthWeeks === 'function') {
        renderMonthWeeks(viewDate.getFullYear(), viewDate.getMonth(), currentStudent);
    }
}

document.addEventListener('DOMContentLoaded', async () => {
    // Fetch students from server first
    await fetchStudents();
    fillStudentSelect();
    renderAll();

    const studentSelect = document.getElementById('studentSelect');
    if (studentSelect) {
        studentSelect.addEventListener('change', (e) => {
            currentStudent = e.target.value;
            renderAll();
        });
    }

    const prevBtn = document.getElementById('prevBtn');
    if (prevBtn) {
        prevBtn.addEventListener('click', () => {
            viewDate.setMonth(viewDate.getMonth() - 1);
            renderAll();
        });
    }

    const nextBtn = document.getElementById('nextBtn');
    if (nextBtn) {
        nextBtn.addEventListener('click', () => {
            viewDate.setMonth(viewDate.getMonth() + 1);
            renderAll();
        });
    }
});
