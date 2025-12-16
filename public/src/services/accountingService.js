/**
 * Accounting Service - Handles API calls for accounting (Buchhaltung) workflow
 * Phase 2: View approved entries and mark as processed (erledigt)
 */
class AccountingService {
    constructor() {
        this.baseUrl = '/api/accounting';
        this.approvedSubmissions = [];
        this.allStudents = [];
        this.studentSubmissions = []; // All submissions for selected student
        this.selectedSubmission = null;
    }

    /**
     * Fetch all approved submissions ready for accounting review
     * @returns {Promise<Array>} Array of approved submissions
     */
    async getApprovedSubmissions() {
        try {
            const response = await fetch(`${this.baseUrl}/approved-entries`, {
                method: 'GET',
                credentials: 'include'
            });

            if (!response.ok) {
                throw new Error('Failed to fetch approved submissions');
            }

            const data = await response.json();
            this.approvedSubmissions = data || [];
            return this.approvedSubmissions;
        } catch (error) {
            console.error('Error fetching approved submissions:', error);
            throw error;
        }
    }

    /**
     * Fetch all students in the system
     * @returns {Promise<Array>} Array of all students
     */
    async getAllStudents() {
        try {
            const response = await fetch(`${this.baseUrl}/students`, {
                method: 'GET',
                credentials: 'include'
            });

            if (!response.ok) {
                throw new Error('Failed to fetch students');
            }

            const data = await response.json();
            this.allStudents = data || [];
            return this.allStudents;
        } catch (error) {
            console.error('Error fetching students:', error);
            throw error;
        }
    }

    /**
     * Get ALL submissions for a specific student (any status)
     * @param {number} studentId - Student ID
     * @returns {Promise<Array>} Array of all submissions for that student
     */
    async getAllSubmissionsForStudent(studentId) {
        try {
            const response = await fetch(`${this.baseUrl}/all-student-submissions?student_id=${studentId}`, {
                method: 'GET',
                credentials: 'include'
            });

            if (!response.ok) {
                throw new Error('Failed to fetch student submissions');
            }

            const data = await response.json();
            this.studentSubmissions = data || [];
            return this.studentSubmissions;
        } catch (error) {
            console.error('Error fetching student submissions:', error);
            throw error;
        }
    }

    /**
     * Get approved submissions for a specific student
     * @param {number} studentId - Student ID
     * @returns {Promise<Array>} Array of approved submissions for that student
     */
    async getApprovedSubmissionsForStudent(studentId) {
        try {
            const response = await fetch(`${this.baseUrl}/student-submissions?student_id=${studentId}`, {
                method: 'GET',
                credentials: 'include'
            });

            if (!response.ok) {
                throw new Error('Failed to fetch student submissions');
            }

            return await response.json();
        } catch (error) {
            console.error('Error fetching student submissions:', error);
            throw error;
        }
    }

    /**
     * Mark a submission as processed (erledigt)
     * @param {string} submissionId - UUID of the submission
     * @returns {Promise<Object>} Result of marking as processed
     */
    async markAsProcessed(submissionId) {
        try {
            const response = await fetch(`${this.baseUrl}/mark-processed`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
                body: JSON.stringify({ submission_id: submissionId })
            });

            if (!response.ok) {
                const errorData = await response.text();
                throw new Error(errorData || 'Failed to mark as processed');
            }

            return await response.json();
        } catch (error) {
            console.error('Error marking submission as processed:', error);
            throw error;
        }
    }

    /**
     * Get time entries for a specific submission
     * @param {number} weekNumber - Week number
     * @param {number} year - Year
     * @param {number} userId - User ID (for accounting to view student's entries)
     * @returns {Promise<Object>} Week data with entries
     */
    async getTimeEntriesForSubmission(weekNumber, year, userId) {
        try {
            const response = await fetch(`/api/time-entries/week?week=${weekNumber}&year=${year}&user_id=${userId}`, {
                method: 'GET',
                credentials: 'include'
            });

            if (!response.ok) {
                throw new Error('Failed to fetch time entries');
            }

            return await response.json();
        } catch (error) {
            console.error('Error fetching time entries:', error);
            throw error;
        }
    }

    /**
     * Get student info by ID
     * @param {number} studentId - Student ID
     * @returns {Object|null} Student object or null
     */
    getStudentById(studentId) {
        return this.allStudents.find(s => s.id === studentId) || null;
    }

    /**
     * Get submissions for a specific student from cached data
     * @param {number} studentId - Student ID
     * @returns {Array} Array of submissions for that student
     */
    getSubmissionsForStudent(studentId) {
        return this.approvedSubmissions.filter(sub => sub.user_id === studentId);
    }

    /**
     * Find submission by ID
     * @param {string} submissionId - UUID of submission
     * @returns {Object|null} Submission object or null
     */
    getSubmissionById(submissionId) {
        return this.approvedSubmissions.find(sub => sub.id === submissionId) || null;
    }
}

// Create global instance
const accountingService = new AccountingService();
