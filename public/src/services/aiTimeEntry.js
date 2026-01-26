// AI Time Entry Service
// Handles natural language parsing of time entries via n8n + Gemini

document.addEventListener('DOMContentLoaded', function() {
    const aiInput = document.getElementById('aiInput');
    const aiParseBtn = document.getElementById('aiParseBtn');
    const aiBtnText = document.getElementById('aiBtnText');
    const aiSpinner = document.getElementById('aiSpinner');
    const aiResult = document.getElementById('aiResult');
    const aiError = document.getElementById('aiError');
    const aiApplyBtn = document.getElementById('aiApplyBtn');
    const aiCancelBtn = document.getElementById('aiCancelBtn');

    // Store parsed data
    let parsedData = null;

    // Parse button click handler
    aiParseBtn.addEventListener('click', async function() {
        const text = aiInput.value.trim();
        
        if (!text) {
            showError('Bitte geben Sie eine Beschreibung ein.');
            return;
        }

        // Show loading state
        setLoading(true);
        hideError();
        hideResult();

        try {
            const response = await fetch('/api/ai/parse-time-entry', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
                body: JSON.stringify({ text: text }),
            });

            const data = await response.json();

            if (!response.ok || !data.success) {
                throw new Error(data.error || 'Fehler beim Verarbeiten');
            }

            parsedData = data.data;
            showResult(parsedData);

        } catch (error) {
            console.error('AI Parse Error:', error);
            showError('Fehler: ' + error.message);
        } finally {
            setLoading(false);
        }
    });

    // Apply button - fills the form with parsed data
    aiApplyBtn.addEventListener('click', function() {
        if (!parsedData) return;

        applyParsedData(parsedData);
        hideResult();
        aiInput.value = '';
        parsedData = null;
    });

    // Cancel button
    aiCancelBtn.addEventListener('click', function() {
        hideResult();
        parsedData = null;
    });

    // Enter key to submit
    aiInput.addEventListener('keydown', function(e) {
        if (e.key === 'Enter' && !e.shiftKey) {
            e.preventDefault();
            aiParseBtn.click();
        }
    });

    function setLoading(loading) {
        aiParseBtn.disabled = loading;
        aiBtnText.textContent = loading ? 'Verarbeite...' : 'Mit KI ausfüllen';
        aiSpinner.classList.toggle('hidden', !loading);
    }

    function showResult(data) {
        document.getElementById('aiDate').textContent = formatDate(data.date);
        
        // Check for special status (Krank, Urlaub, Feiertag)
        const specialStatus = getSpecialStatus(data);
        const specialRow = document.getElementById('aiSpecialRow');
        
        if (specialStatus) {
            document.getElementById('aiDuration').textContent = '-';
            document.getElementById('aiSpecial').textContent = specialStatus;
            specialRow.classList.remove('hidden');
        } else {
            // Show time range if available, otherwise show duration
            if (data.startTime && data.endTime) {
                document.getElementById('aiDuration').textContent = `${data.startTime} - ${data.endTime} (${formatDuration(data.durationMinutes)})`;
            } else {
                document.getElementById('aiDuration').textContent = formatDuration(data.durationMinutes);
            }
            specialRow.classList.add('hidden');
        }
        
        aiResult.classList.remove('hidden');
    }
    
    function getSpecialStatus(data) {
        const text = (data.activity || data.description || '').toLowerCase();
        if (text.includes('krank') || text.includes('sick')) return 'Krank';
        if (text.includes('urlaub') || text.includes('vacation') || text.includes('holiday')) return 'Urlaub';
        if (text.includes('feiertag') || text.includes('public holiday')) return 'Feiertag';
        return null;
    }

    function hideResult() {
        aiResult.classList.add('hidden');
    }

    function showError(message) {
        aiError.textContent = message;
        aiError.classList.remove('hidden');
    }

    function hideError() {
        aiError.classList.add('hidden');
    }

    function formatDate(dateStr) {
        if (!dateStr) return '-';
        const date = new Date(dateStr);
        const day = date.getDate().toString().padStart(2, '0');
        const month = (date.getMonth() + 1).toString().padStart(2, '0');
        return `${day}.${month}`;
    }

    function formatDuration(minutes) {
        if (!minutes) return '-';
        const hours = Math.floor(minutes / 60);
        const mins = minutes % 60;
        if (mins === 0) {
            return `${hours} Stunde${hours !== 1 ? 'n' : ''}`;
        }
        return `${hours}h ${mins}min`;
    }

    function applyParsedData(data) {
        // Find the correct row based on the date
        const targetDate = new Date(data.date);
        const dayOfWeek = targetDate.getDay(); // 0 = Sunday, 1 = Monday, etc.
        
        // Map day of week to row index (Monday = 0, Sunday = 6)
        const rowIndex = dayOfWeek === 0 ? 6 : dayOfWeek - 1;
        
        // Get the table rows (excluding header and total row)
        const table = document.getElementById('timeTable');
        const rows = table.querySelectorAll('tbody tr:not(.total-row)');
        
        if (rowIndex < 0 || rowIndex >= rows.length) {
            showError('Das Datum liegt außerhalb der aktuellen Woche.');
            return;
        }

        const row = rows[rowIndex];
        
        // Check if this is a weekend row (disabled)
        if (rowIndex >= 5) {
            showError('Wochenenden können nicht bearbeitet werden.');
            return;
        }

        const notesInput = row.querySelector('.notes-input');
        const timeInputs = row.querySelectorAll('.time-input');
        
        // Check for special status (Krank, Urlaub, Feiertag)
        const specialStatus = getSpecialStatus(data);
        
        if (specialStatus) {
            // For special cases, only fill the notes field
            if (notesInput) {
                notesInput.value = specialStatus;
            }
        } else {
            // Normal time entry - fill times only
            let startTime, endTime;
            
            if (data.startTime && data.endTime) {
                startTime = data.startTime;
                endTime = data.endTime;
            } else if (data.durationMinutes > 0) {
                const durationHours = data.durationMinutes / 60;
                const startHour = 8;
                const endHour = startHour + durationHours;
                startTime = `${String(startHour).padStart(2, '0')}:00`;
                endTime = `${String(Math.floor(endHour)).padStart(2, '0')}:${String(Math.round((endHour % 1) * 60)).padStart(2, '0')}`;
            }
            
            if (timeInputs.length >= 2 && startTime && endTime) {
                timeInputs[0].value = startTime;
                timeInputs[1].value = endTime;
                
                timeInputs[0].dispatchEvent(new Event('change', { bubbles: true }));
                timeInputs[1].dispatchEvent(new Event('change', { bubbles: true }));
            }
        }

        // Show success feedback
        row.style.backgroundColor = '#e8f5e9';
        setTimeout(() => {
            row.style.backgroundColor = '';
        }, 2000);
    }
});
