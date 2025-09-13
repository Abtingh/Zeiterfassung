document.addEventListener('DOMContentLoaded', function() {
    const resetPasswordLink = document.getElementById('resetPasswordLink');

    resetPasswordLink.addEventListener('click', function(e) {
        e.preventDefault();
        console.log('Reset Password clicked');
        window.location.href = '/reset-passwort';
    });
});
