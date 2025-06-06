document.addEventListener('DOMContentLoaded', function () {
// This script handles the logout functionality by sending a POST request to the server
document.getElementById('logoutBtn').onclick = async function() {
    // Get email from somewhere (e.g., a JS variable or hidden field)
    alert('Button clicked!');
    const email = localStorage.getItem('userEmail'); // or however you store it

    // Get CSRF token from cookie (if you store it as a cookie)
    function getCookie(name) {
        const value = `; ${document.cookie}`;
        const parts = value.split(`; ${name}=`);
        if (parts.length === 2) return parts.pop().split(';').shift();
    }
    const csrfToken = getCookie('csrf_token');

    // Prepare form data
    const formData = new FormData();
    formData.append('email', email);

    // Send POST request to /logout
    const response = await fetch('/logout', {
        method: 'POST',
        body: formData,
        headers: {
            'X-CSRF-Token': csrfToken
        },
        credentials: 'include' // Important: send cookies
    });

    if (response.ok) {
        alert('Logged out!');
        window.location.href = '/login'; // Redirect to login page
    } else {
        alert('Logout failed!');
    }
};
});