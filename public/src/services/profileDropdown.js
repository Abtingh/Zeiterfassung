document.addEventListener('DOMContentLoaded', function() {
    const profileIcon = document.getElementById('iconProfile');
    const profileImg = document.querySelector('#iconProfile img'); // Get the img element
    const dropdown = document.getElementById('profileDropDown');
    

    // Toggle dropdown when clicking on profile icon
    profileIcon.addEventListener('click', function(e) {
        // If the click happened inside the dropdown, do nothing
        if (e.target.closest('#profileDropDown')) return;

        e.preventDefault();
        e.stopPropagation(); // Prevent event bubbling 

        // Toggle the clicked class on the img element
        profileImg.classList.toggle('clicked');

        dropdown.classList.toggle('hidden');
        dropdown.classList.toggle('show');
    });

    // Close dropdown when clicking outside
    document.addEventListener('click', function(e) {
        if (!dropdown.contains(e.target) && !profileIcon.contains(e.target)) {
            dropdown.classList.add('hidden');
            dropdown.classList.remove('show');
            // Remove clicked state when dropdown closes
            profileImg.classList.remove('clicked');
        }
    });
    
    // Handle menu item clicks
    document.getElementById('profileLink').addEventListener('click', function(e) {
        e.preventDefault();
        console.log('Profile clicked');
        // Add your profile logic here
        closeDropdown();
    });
    
    document.getElementById('resetPasswordLink').addEventListener('click', function(e) {
        e.preventDefault();
        console.log('Reset Password clicked');
        window.location.href = '/reset-passwort';
        closeDropdown();
    });
    
    document.getElementById('logoutLink').addEventListener('click', function(e) {
        e.preventDefault();
        console.log('Logout clicked');
        window.location.href = '/logout';
        closeDropdown();
    });
    
    // Helper function to close dropdown
    function closeDropdown() {
        dropdown.classList.add('hidden');
        dropdown.classList.remove('show');
        // Remove clicked state when dropdown closes
        profileImg.classList.remove('clicked');
    }
});
