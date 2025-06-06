document.addEventListener('DOMContentLoaded', function () {
    const form = document.querySelector('form');
    form.addEventListener('submit', async function (e) {
        e.preventDefault();

        const email = document.getElementById('Email').value;
        const password = document.getElementById('Password').value;

        const formData = new FormData();
        formData.append('email', email);
        formData.append('password', password);

        const response = await fetch('/login', {
            method: 'POST',
            body: formData,
            credentials: 'include'
        });

        // Try to store email for logout
        if (response.ok) {
            localStorage.setItem('userEmail', email);

            // If the response is HTML, redirect to the new page
            const contentType = response.headers.get('content-type');
            if (contentType && contentType.includes('text/html')) {
                // Replace the current page with the response HTML
                const html = await response.text();
                document.open();
                document.write(html);
                document.close();
            } else {
                // Otherwise, handle as text (e.g., error or message)
                const msg = await response.text();
                alert(msg);
            }
        } else {
            const msg = await response.text();
            alert(msg || 'Login failed!');
        }
    });
});