// Fetch user information and update the UI
async function loadUserInfo() {
    try {
        const response = await fetch('/me', {
            method: 'GET',
            credentials: 'include', // Include cookies for session
            headers: {
                'Content-Type': 'application/json'
            }
        });

        if (response.ok) {
            const userData = await response.json();
            updateUserDisplay(userData.first_name);
        } else if (response.status === 401) {
            // User not authenticated, redirect to login
            window.location.href = '/login';
        } else {
            console.error('Failed to fetch user info:', response.statusText);
            // Fallback display
            updateUserDisplay('User');
        }
    } catch (error) {
        console.error('Error fetching user info:', error);
        // Fallback display
        updateUserDisplay('User');
    }
}

// Update the name display in the UI
function updateUserDisplay(firstName) {
    const nameHolder = document.querySelector('.nameHolder');
    if (nameHolder) {
        nameHolder.textContent = `Hi, ${firstName}`;
    }
}

// Load user info when the page loads
document.addEventListener('DOMContentLoaded', loadUserInfo);