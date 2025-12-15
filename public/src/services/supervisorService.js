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
            const response = await fetch(`${this.baseUrl}/pending-entries`, {
                method: 'GET',
                credentials: 'include'
            });

            if (!response.ok) {
                throw new Error('Failed to fetch pending submissions');
            }

            const data = await response.json();
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
                studentMap.set(sub.user_id, {
                    id: sub.user_id,
                    name: `${sub.first_name?.String || ''} ${sub.last_name?.String || ''}`.trim(),
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
}

// Create global instance
const supervisorService = new SupervisorService();
