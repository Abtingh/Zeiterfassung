document.addEventListener('DOMContentLoaded', function () {
    const form = document.querySelector('form');
    if (!form) {
        console.error('Form not found!');
        return;
    }
    
    // Simply let the form submit naturally - no JavaScript fetch
    // The backend will handle the redirect after successful authentication
    form.addEventListener('submit', function(e) {
        const email = document.getElementById('Email').value;
        const password = document.getElementById('Password').value;
        
        if (!email || !password) {
            e.preventDefault();
            alert('Please fill in both email and password');
            return;
        }
        
        // Store email in localStorage for potential use later
        localStorage.setItem('userEmail', email);
        
        // Let the form submit naturally - the backend redirect will work
        console.log('Submitting login form for:', email);
    });
});