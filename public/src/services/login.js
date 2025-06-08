document.addEventListener('DOMContentLoaded', function () {
    const form = document.querySelector('form');
    if (!form) {
        console.error('Form not found!');
        return;
    }
    
    form.addEventListener('submit', async function(e) {
        e.preventDefault();
        
        const email = document.getElementById('Email').value;
        const password = document.getElementById('Password').value;
        
        const params = new URLSearchParams();
        params.append('email', email);
        params.append('password', password);
        
        try {
            const response = await fetch('/login', {
                method: 'POST',
                body: params,
                headers: {
                    'Content-Type': 'application/x-www-form-urlencoded'
                },
                credentials: 'include'
            });

            const data = await response.json();
            console.log("Response status:", response.status);
            console.log("Response data:", data);

            if (response.ok && data.redirect) {
                localStorage.setItem('userEmail', email);
                console.log("Redirecting to:", data.redirect);
                window.location.href = data.redirect;
            } else {
                const errorMsg = data.error || 'Login failed!';
                alert(errorMsg);
            }
        } catch (err) {
            console.error('Login error:', err);
            alert('Network error. Please try again.');
        }
    });
});