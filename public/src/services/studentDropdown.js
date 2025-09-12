// داده‌ی نمونهٔ دانشجوها — بعداً از سرور بگیر
const students = [
  { id: "s1", name: "Student 1" },
  { id: "s2", name: "Student 2" },
  { id: "s3", name: "Student 3" },
];

const monthFmt = new Intl.DateTimeFormat('de-DE',{month:'long', year:'numeric'});
let viewDate = new Date();       // ماه فعلی
let currentStudent = "s1";       // پیش‌فرض

function fillStudentSelect(){
  const sel = document.getElementById('studentSelect');
  sel.innerHTML = students.map(s => `<option value="${s.id}">${s.name}</option>`).join('');
  sel.value = currentStudent;
}

function renderHeader(){
  document.getElementById('monthLabel').textContent = monthFmt.format(viewDate);
}

function renderAll(){
  renderHeader();
  // اگر وضعیت‌ها به دانشجو وابسته‌اند، اینجا studentId را پاس بده
  renderMonthWeeks(viewDate.getFullYear(), viewDate.getMonth(), currentStudent);
}

document.addEventListener('DOMContentLoaded', () => {
  fillStudentSelect();
  renderAll();

  document.getElementById('studentSelect').addEventListener('change', (e)=>{
    currentStudent = e.target.value;
    renderAll();
  });

  document.getElementById('prevBtn').addEventListener('click', ()=>{
    viewDate.setMonth(viewDate.getMonth() - 1);
    renderAll();
  });
  document.getElementById('nextBtn').addEventListener('click', ()=>{
    viewDate.setMonth(viewDate.getMonth() + 1);
    renderAll();
  });
});
