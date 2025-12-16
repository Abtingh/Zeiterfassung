/**
 * Supervisor Service - Handles API calls for supervisor workflow
 * Phase 1: Approve/Reject time entry submissions
 */
class SupervisorService {
    constructor() {
        this.baseUrl = '/api/supervisor';
        this.pendingSubmissions = [];
        this.selectedSubmission = null;
    }

    /**
     * Fetch all pending submissions for this supervisor
     * @returns {Promise<Array>} Array of pending submissions
     */
    async getPendingSubmissions() {
        try {
            console.log('Fetching pending submissions from:', `${this.baseUrl}/pending-entries`);
            
            const response = await fetch(`${this.baseUrl}/pending-entries`, {
                method: 'GET',
                credentials: 'include'
            });

            console.log('Response status:', response.status);
            
            if (!response.ok) {
                const errorText = await response.text();
                console.error('API error response:', errorText);
                throw new Error('Failed to fetch pending submissions');
            }

            const data = await response.json();
            console.log('Raw API response:', data);
            
            this.pendingSubmissions = data || [];
            return this.pendingSubmissions;
        } catch (error) {
            console.error('Error fetching pending submissions:', error);
            throw error;
        }
    }

    /**
     * Approve a submission
     * @param {string} submissionId - UUID of the submission
     * @returns {Promise<Object>} Result of the approval
     */
    async approveSubmission(submissionId) {
        try {
            const response = await fetch(`${this.baseUrl}/approve`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
                body: JSON.stringify({ submission_id: submissionId })
            });

            if (!response.ok) {
                const errorData = await response.text();
                throw new Error(errorData || 'Failed to approve submission');
            }

            return await response.json();
        } catch (error) {
            console.error('Error approving submission:', error);
            throw error;
        }
    }

    /**
     * Reject a submission
     * @param {string} submissionId - UUID of the submission
     * @param {string} reason - Reason for rejection
     * @returns {Promise<Object>} Result of the rejection
     */
    async rejectSubmission(submissionId, reason = '') {
        try {
            const response = await fetch(`${this.baseUrl}/reject`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
                body: JSON.stringify({ 
                    submission_id: submissionId,
                    reason: reason
                })
            });

            if (!response.ok) {
                const errorData = await response.text();
                throw new Error(errorData || 'Failed to reject submission');
            }

            return await response.json();
        } catch (error) {
            console.error('Error rejecting submission:', error);
            throw error;
        }
    }

    /**
     * Get unique students from pending submissions
     * @returns {Array} Array of unique students
     */
    getUniqueStudents() {
        const studentMap = new Map();
        this.pendingSubmissions.forEach(sub => {
            if (!studentMap.has(sub.user_id)) {
                // Handle different formats for first_name/last_name
                let firstName = '';
                let lastName = '';
                
                if (sub.first_name) {
                    firstName = typeof sub.first_name === 'object' ? (sub.first_name.String || '') : sub.first_name;
                }
                if (sub.last_name) {
                    lastName = typeof sub.last_name === 'object' ? (sub.last_name.String || '') : sub.last_name;
                }
                
                const fullName = `${firstName} ${lastName}`.trim();
                
                studentMap.set(sub.user_id, {
                    id: sub.user_id,
                    name: fullName || sub.email,
                    email: sub.email
                });
            }
        });
        return Array.from(studentMap.values());
    }

    /**
     * Get submissions for a specific student
     * @param {number} userId - User ID of the student
     * @returns {Array} Array of submissions for that student
     */
    getSubmissionsForStudent(userId) {
        return this.pendingSubmissions.filter(sub => sub.user_id === userId);
    }

    /**
     * Find submission by ID
     * @param {string} submissionId - UUID of submission
     * @returns {Object|null} Submission object or null
     */
    getSubmissionById(submissionId) {
        return this.pendingSubmissions.find(sub => sub.id === submissionId) || null;
    }

    /**
     * Get time entries for a specific submission
     * @param {string} submissionId - UUID of the submission
     * @returns {Promise<Array>} Array of time entries
     */
    async getTimeEntriesForSubmission(submissionId) {
        try {
            console.log('Fetching time entries for submission:', submissionId);
            
            const response = await fetch(`${this.baseUrl}/submission-entries?submission_id=${submissionId}`, {
                method: 'GET',
                credentials: 'include'
            });

            console.log('Response status:', response.status);
            
            if (!response.ok) {
                const errorText = await response.text();
                console.error('API error response:', errorText);
                throw new Error('Failed to fetch time entries for submission');
            }

            const data = await response.json();
            console.log('Time entries for submission:', data);
            
            return data || [];
        } catch (error) {
            console.error('Error fetching time entries for submission:', error);
            throw error;
        }
    }
}

// Create global instance
const supervisorService = new SupervisorService();
