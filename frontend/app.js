const { createApp, ref, onMounted } = Vue

createApp({
    setup() {
        const isLoggedIn = ref(false)
        const loading = ref(false)
        const loadingDestinations = ref(false)
        const loginError = ref('')
        const loginData = ref({ username: '', password: '' })
        const destinations = ref([])

        const showAddModal = ref(false)
        const isEdit = ref(false)
        const formData = ref({
            ID: null,
            AppName: '',
            RoutingCode: '',
            TargetURL: '',
            SecretToken: '',
            ProviderToken: ''
        })

        const API_BASE = '/api'

        // Check auth status on load
        onMounted(() => {
            const token = localStorage.getItem('token')
            if (token) {
                isLoggedIn.value = true
                fetchDestinations()
            }
        })

        const getHeaders = () => {
            return {
                'Content-Type': 'application/json',
                'Authorization': 'Bearer ' + localStorage.getItem('token')
            }
        }

        const login = async () => {
            loading.value = true
            loginError.value = ''
            try {
                const res = await fetch(`${API_BASE}/login`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(loginData.value)
                })
                const data = await res.json()
                
                if (!res.ok) throw new Error(data.message || 'Login failed')
                
                localStorage.setItem('token', data.token)
                isLoggedIn.value = true
                loginData.value = { username: '', password: '' }
                fetchDestinations()
            } catch (err) {
                loginError.value = err.message
            } finally {
                loading.value = false
            }
        }

        const logout = () => {
            localStorage.removeItem('token')
            isLoggedIn.value = false
            destinations.value = []
        }

        const fetchDestinations = async () => {
            loadingDestinations.value = true
            try {
                const res = await fetch(`${API_BASE}/destinations`, { headers: getHeaders() })
                if (res.status === 401) {
                    logout()
                    return
                }
                const data = await res.json()
                destinations.value = data.data || []
            } catch (err) {
                console.error("Failed to fetch destinations", err)
            } finally {
                loadingDestinations.value = false
            }
        }

        const closeModal = () => {
            showAddModal.value = false
            isEdit.value = false
            formData.value = {
                ID: null,
                AppName: '',
                RoutingCode: '',
                TargetURL: '',
                SecretToken: '',
                ProviderToken: ''
            }
        }

        const editDest = (dest) => {
            isEdit.value = true
            formData.value = { ...dest }
            showAddModal.value = true
        }

        const deleteDest = async (id) => {
            if (!confirm('Apakah Anda yakin ingin menghapus destination ini?')) return
            
            try {
                await fetch(`${API_BASE}/destinations/${id}`, {
                    method: 'DELETE',
                    headers: getHeaders()
                })
                fetchDestinations()
            } catch (err) {
                console.error(err)
            }
        }

        const saveDestination = async () => {
            loading.value = true
            try {
                const method = isEdit.value ? 'PUT' : 'POST'
                const url = isEdit.value 
                    ? `${API_BASE}/destinations/${formData.value.ID}` 
                    : `${API_BASE}/destinations`

                const res = await fetch(url, {
                    method,
                    headers: getHeaders(),
                    body: JSON.stringify(formData.value)
                })

                if (!res.ok) {
                    const data = await res.json()
                    throw new Error(data.message || 'Failed to save')
                }

                closeModal()
                fetchDestinations()
            } catch (err) {
                alert(err.message)
            } finally {
                loading.value = false
            }
        }

        const generateSecret = () => {
            const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789'
            let token = ''
            for (let i = 0; i < 32; i++) {
                token += chars.charAt(Math.floor(Math.random() * chars.length))
            }
            formData.value.SecretToken = token
        }

        return {
            isLoggedIn, loading, loadingDestinations, loginError, loginData, destinations,
            showAddModal, isEdit, formData,
            login, logout, editDest, deleteDest, saveDestination, closeModal, generateSecret
        }
    }
}).mount('#app')
