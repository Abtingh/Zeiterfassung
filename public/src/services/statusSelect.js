// Status Select Handler
// Disables time fields when a special status (Krank, Urlaub, Feiertag, Frei) is selected

document.addEventListener('DOMContentLoaded', function() {
    // Get all status selects
    const statusSelects = document.querySelectorAll('.status-select');
    
    statusSelects.forEach(select => {
        select.addEventListener('change', function() {
            handleStatusChange(this);
        });
    });
    
    function handleStatusChange(selectElement) {
        const row = selectElement.closest('tr');
        const timeInputs = row.querySelectorAll('.time-input');
        const pauseInput = row.querySelector('.pause-input');
        const calculatedHours = row.querySelector('.calculated-hours');
        const workHours = row.querySelector('.work-hours');
        
        if (selectElement.value !== '') {
            // Special status selected - disable time fields and clear them
            timeInputs.forEach(input => {
                input.disabled = true;
                input.value = '';
            });
            if (pauseInput) {
                pauseInput.disabled = true;
                pauseInput.value = '';
            }
            if (calculatedHours) calculatedHours.textContent = '-';
            if (workHours) workHours.textContent = '-';
            
            // Add visual indicator
            row.classList.add('special-status-row');
        } else {
            // No status - enable time fields
            timeInputs.forEach(input => {
                input.disabled = false;
            });
            if (pauseInput) {
                pauseInput.disabled = false;
            }
            
            // Remove visual indicator
            row.classList.remove('special-status-row');
        }
    }
});
