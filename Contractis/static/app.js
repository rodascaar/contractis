// DOM elements
const uploadArea = document.getElementById('uploadArea');
const fileInput = document.getElementById('fileInput');
const selectFileBtn = document.getElementById('selectFileBtn');
const fileInfo = document.getElementById('fileInfo');
const fileName = document.getElementById('fileName');
const analyzeBtn = document.getElementById('analyzeBtn');
const resultsSection = document.getElementById('resultsSection');
const resultsContent = document.getElementById('resultsContent');
const loadingSection = document.getElementById('loadingSection');
const loadingProgress = document.getElementById('loadingProgress');
const errorSection = document.getElementById('errorSection');
const errorMessage = document.getElementById('errorMessage');
const downloadBtn = document.getElementById('downloadBtn');

// Estimation elements
const estimateBtn = document.getElementById('estimateBtn');
const estimationSection = document.getElementById('estimationSection');
const proceedAnalyzeBtn = document.getElementById('proceedAnalyzeBtn');
const cancelEstimateBtn = document.getElementById('cancelEstimateBtn');
const applyRecommendedBtn = document.getElementById('applyRecommendedBtn');

let currentEstimation = null;

// Settings modal elements
const settingsBtn = document.getElementById('settingsBtn');
const settingsModal = document.getElementById('settingsModal');
const closeModal = document.getElementById('closeModal');
const cancelBtn = document.getElementById('cancelBtn');
const saveBtn = document.getElementById('saveBtn');
const llmType = document.getElementById('llmType');
const localSettings = document.getElementById('localSettings');
const onlineSettings = document.getElementById('onlineSettings');
const localUrl = document.getElementById('localUrl');
const apiUrl = document.getElementById('apiUrl');
const apiKey = document.getElementById('apiKey');
// Note: modelName is now split into localModelName and onlineModelName
// We'll get the appropriate one dynamically in saveSettings()
const maxTokens = document.getElementById('maxTokens');

// History modal elements
const historyBtn = document.getElementById('historyBtn');
const historyModal = document.getElementById('historyModal');
const closeHistoryModal = document.getElementById('closeHistoryModal');
const historySearch = document.getElementById('historySearch');
const refreshHistoryBtn = document.getElementById('refreshHistoryBtn');
const historyList = document.getElementById('historyList');

// File handling
let selectedFile = null;

// Load initial LLM configuration from localStorage or use defaults
function loadInitialConfig() {
    const saved = localStorage.getItem('llmConfig');
    if (saved) {
        try {
            const config = JSON.parse(saved);
            return config;
        } catch (e) {
            console.error('Error parsing saved config:', e);
        }
    }
    // Default configuration
    return {
        type: 'local',
        localUrl: 'http://localhost:1234/v1/chat/completions',
        apiUrl: '',
        apiKey: '',
        modelName: '',
        maxTokens: 800
    };
}

// LLM Configuration - Load from localStorage or use defaults
let llmConfig = loadInitialConfig();

// Update LLM status indicator
function updateLLMStatusIndicator() {
    const indicator = document.getElementById('llmIndicator');
    if (indicator) {
        if (llmConfig.type === 'online') {
            indicator.textContent = '🔵 En línea';
            indicator.style.background = 'rgba(59, 130, 246, 0.2)';
        } else {
            indicator.textContent = '🟢 Local';
            indicator.style.background = 'rgba(34, 197, 94, 0.2)';
        }
    }
}

// Event listeners
selectFileBtn.addEventListener('click', (e) => {
    e.stopPropagation();
    fileInput.click();
});

fileInput.addEventListener('change', handleFileSelect);

uploadArea.addEventListener('dragover', (e) => {
    e.preventDefault();
    uploadArea.classList.add('dragover');
});

uploadArea.addEventListener('dragleave', () => {
    uploadArea.classList.remove('dragover');
});

uploadArea.addEventListener('drop', (e) => {
    e.preventDefault();
    uploadArea.classList.remove('dragover');
    const files = e.dataTransfer.files;
    if (files.length > 0) {
        handleFile(files[0]);
    }
});

uploadArea.addEventListener('click', (e) => {
    if (e.target === uploadArea || e.target === uploadArea.querySelector('.upload-icon') || e.target === uploadArea.querySelector('h3') || e.target === uploadArea.querySelector('p')) {
        fileInput.click();
    }
});

estimateBtn.addEventListener('click', estimateTokens);
analyzeBtn.addEventListener('click', analyzeContract);
proceedAnalyzeBtn.addEventListener('click', () => {
    hideEstimation();
    analyzeContract();
});
cancelEstimateBtn.addEventListener('click', () => {
    hideEstimation();
    estimateBtn.style.display = 'inline-block';
    analyzeBtn.style.display = 'none';
});
applyRecommendedBtn.addEventListener('click', applyRecommendedConfig);

downloadBtn.addEventListener('click', downloadPDF);

// Settings modal event listeners
settingsBtn.addEventListener('click', openSettingsModal);
closeModal.addEventListener('click', closeSettingsModal);
cancelBtn.addEventListener('click', closeSettingsModal);
saveBtn.addEventListener('click', saveSettings);
llmType.addEventListener('change', toggleLLMSettings);

// History modal event listeners
historyBtn.addEventListener('click', openHistoryModal);
closeHistoryModal.addEventListener('click', closeHistoryModalFn);
refreshHistoryBtn.addEventListener('click', loadHistory);
historySearch.addEventListener('input', debounce(searchHistory, 500));

// Close modal when clicking outside
window.addEventListener('click', (e) => {
    if (e.target === settingsModal) {
        closeSettingsModal();
    }
    if (e.target === historyModal) {
        closeHistoryModalFn();
    }
});

// Functions
function handleFileSelect(e) {
    const file = e.target.files[0];
    if (file) {
        handleFile(file);
    }
}

function handleFile(file) {
    if (!file.type.includes('pdf')) {
        showError('Solo se permiten archivos PDF');
        return;
    }

    selectedFile = file;
    fileName.textContent = file.name;
    fileInfo.style.display = 'block';
    hideError();
}

function estimateTokens() {
    if (!selectedFile) {
        showError('Por favor selecciona un archivo PDF');
        return;
    }

    hideResults();
    hideError();
    hideEstimation();
    showLoading();
    updateLoadingProgress('Analizando documento y estimando tokens...');

    const formData = new FormData();
    formData.append('file', selectedFile);
    formData.append('maxTokens', llmConfig.maxTokens.toString());

    fetch('/estimate', {
        method: 'POST',
        body: formData
    })
    .then(response => response.json())
    .then(data => {
        hideLoading();
        if (data.success) {
            currentEstimation = data;
            showEstimation(data);
        } else {
            showError(data.error || 'Error al estimar tokens');
        }
    })
    .catch(error => {
        hideLoading();
        showError('Error de conexión: ' + error.message);
    });
}

function showEstimation(data) {
    // Actualizar valores
    document.getElementById('estCharacters').textContent = data.characterCount.toLocaleString();
    document.getElementById('estTokens').textContent = data.estimatedTokens.toLocaleString();
    document.getElementById('estChunks').textContent = data.chunks;
    document.getElementById('estTotal').textContent = data.totalTokens.toLocaleString();
    
    document.getElementById('estSystemPrompt').textContent = data.systemPromptTokens.toLocaleString() + ' tokens';
    document.getElementById('estPhase1').textContent = data.phase1Tokens.toLocaleString() + ' tokens';
    document.getElementById('estPhase2Input').textContent = data.phase2InputTokens.toLocaleString() + ' tokens';
    document.getElementById('estPhase2Output').textContent = data.phase2OutputTokens.toLocaleString() + ' tokens';
    document.getElementById('estTotalDetail').textContent = data.totalTokens.toLocaleString() + ' tokens';
    
    document.getElementById('estRecommended').textContent = data.recommendedMaxTokens;
    document.getElementById('estCurrentConfig').innerHTML = 
        'Tu configuración actual: <strong>' + llmConfig.maxTokens + '</strong>';
    
    // Mostrar botón de aplicar si la recomendación es diferente
    if (data.recommendedMaxTokens !== llmConfig.maxTokens) {
        applyRecommendedBtn.style.display = 'inline-block';
    } else {
        applyRecommendedBtn.style.display = 'none';
    }
    
    // Mostrar advertencia si existe
    if (data.warning) {
        document.getElementById('estWarningText').textContent = '⚠️ ' + data.warning;
        document.getElementById('estWarning').style.display = 'block';
    } else {
        document.getElementById('estWarning').style.display = 'none';
    }
    
    estimationSection.style.display = 'block';
    estimateBtn.style.display = 'none';
    analyzeBtn.style.display = 'inline-block';
}

function hideEstimation() {
    estimationSection.style.display = 'none';
}

function applyRecommendedConfig() {
    if (currentEstimation) {
        llmConfig.maxTokens = currentEstimation.recommendedMaxTokens;
        maxTokens.value = currentEstimation.recommendedMaxTokens;
        localStorage.setItem('llmConfig', JSON.stringify(llmConfig));
        
        // Actualizar UI
        document.getElementById('estCurrentConfig').innerHTML =
            'Tu configuración actual: <strong>' + llmConfig.maxTokens + '</strong>';
        applyRecommendedBtn.style.display = 'none';

        showSuccess('✓ Configuración actualizada a ' + currentEstimation.recommendedMaxTokens + ' tokens', 3000);
    }
}

function analyzeContract() {
    if (!selectedFile) {
        showError('Por favor selecciona un archivo PDF');
        return;
    }

    showLoading();
    hideResults();
    hideError();
    hideEstimation();

    const formData = new FormData();
    formData.append('file', selectedFile);
    formData.append('llmConfig', JSON.stringify(llmConfig));

    fetch('/upload', {
        method: 'POST',
        body: formData,
        // Evitar timeout automático del navegador para modelos locales
        signal: llmConfig.type === 'local' ? undefined : AbortSignal.timeout(60000) // 60s para online, sin límite para local
    })
    .then(response => {
        if (!response.ok) {
            throw new Error(`HTTP ${response.status}: ${response.statusText}`);
        }
        return response.json();
    })
    .then(data => {
        hideLoading();
        if (data.success) {
            showResults(data.data);
        } else {
            console.error('❌ Error en respuesta del backend:', data.error);
            showError(data.error || 'Error desconocido');
        }
    })
    .catch(error => {
        hideLoading();
        console.error('❌ Error de conexión en /upload:', error);

        // Mensajes de error más específicos
        let errorMessage = 'Error de conexión';
        if (error.name === 'AbortError') {
            errorMessage = 'La solicitud fue cancelada por timeout';
        } else if (error.message.includes('Failed to fetch')) {
            errorMessage = 'No se pudo conectar al servidor. Verifica que esté ejecutándose.';
        } else if (error.message.includes('HTTP')) {
            errorMessage = error.message;
        } else {
            errorMessage = error.message || 'Error desconocido';
        }

        showError(errorMessage);
    });
}

function showResults(content, contractInfo = null) {
    // Validar que content sea válido
    if (content === null || content === undefined) {
        console.error('showResults recibió content null/undefined');
        showError('Error: No se recibió contenido del análisis');
        return;
    }

    // Formatear el contenido para mejor legibilidad
    const formattedContent = formatAnalysisContent(content);

    // Agregar información del contrato si está disponible
    let infoHtml = '';
    if (contractInfo && contractInfo.filename) {
        const date = contractInfo.uploaded_at ? new Date(contractInfo.uploaded_at).toLocaleString('es-ES') : 'N/A';
        const llmType = contractInfo.llm_type === 'online' ? '🔵 En línea' : '🟢 Local';
        const model = contractInfo.llm_model || 'N/A';

        infoHtml = `
            <div style="background: #f8f9fa; border-left: 4px solid #4CAF50; padding: 15px; margin-bottom: 20px; border-radius: 4px;">
                <div style="font-size: 1.1em; font-weight: bold; margin-bottom: 10px; color: #333;">
                    📄 ${escapeHtml(contractInfo.filename)}
                </div>
                <div style="font-size: 0.9em; color: #666; display: flex; gap: 20px; flex-wrap: wrap;">
                    <span><strong>Fecha:</strong> ${date}</span>
                    <span><strong>Modelo:</strong> ${llmType} ${escapeHtml(model)}</span>
                </div>
            </div>
        `;
    }

    resultsContent.innerHTML = infoHtml + formattedContent;

    // Actualizar encabezado
    const resultsHeader = document.querySelector('.results-header h3');
    resultsHeader.textContent = 'Resultado del Análisis';

    resultsSection.style.display = 'block';
}

function hideResults() {
    resultsSection.style.display = 'none';
}

function showLoading() {
    loadingSection.style.display = 'block';
    updateLoadingProgress('Procesando documento...');
    disableControls();
}

function hideLoading() {
    loadingSection.style.display = 'none';
    enableControls();
}

function disableControls() {
    // Deshabilitar área de carga
    uploadArea.style.pointerEvents = 'none';
    uploadArea.style.opacity = '0.5';
    fileInput.disabled = true;
    
    // Deshabilitar botones
    selectFileBtn.disabled = true;
    estimateBtn.disabled = true;
    analyzeBtn.disabled = true;
    proceedAnalyzeBtn.disabled = true;
    settingsBtn.disabled = true;
    
    // Agregar clase visual
    selectFileBtn.classList.add('disabled');
    estimateBtn.classList.add('disabled');
    analyzeBtn.classList.add('disabled');
    proceedAnalyzeBtn.classList.add('disabled');
    settingsBtn.classList.add('disabled');
}

function enableControls() {
    // Habilitar área de carga
    uploadArea.style.pointerEvents = 'auto';
    uploadArea.style.opacity = '1';
    fileInput.disabled = false;
    
    // Habilitar botones
    selectFileBtn.disabled = false;
    estimateBtn.disabled = false;
    analyzeBtn.disabled = false;
    proceedAnalyzeBtn.disabled = false;
    settingsBtn.disabled = false;
    
    // Remover clase visual
    selectFileBtn.classList.remove('disabled');
    estimateBtn.classList.remove('disabled');
    analyzeBtn.classList.remove('disabled');
    proceedAnalyzeBtn.classList.remove('disabled');
    settingsBtn.classList.remove('disabled');
}

function updateLoadingProgress(message) {
    if (loadingProgress) {
        loadingProgress.textContent = message;
    }
}

function showError(message) {
    errorMessage.textContent = message;
    errorSection.style.display = 'block';

    // Asegurar que sea visible
    errorSection.scrollIntoView({ behavior: 'smooth', block: 'center' });
}

function hideError() {
    errorSection.style.display = 'none';
}

// Función adicional para mostrar mensajes de éxito temporales
function showSuccess(message, duration = 3000) {
    showError(message); // Reutilizar showError para consistencia
    setTimeout(() => {
        hideError();
    }, duration);
}

function formatAnalysisContent(content) {
    // Validar que content sea una cadena
    if (typeof content !== 'string' || content === null || content === undefined) {
        console.error('formatAnalysisContent recibió un valor no válido:', content);
        return '<p>Error: Contenido de análisis no válido</p>';
    }

    try {
        // Convertir saltos de línea en párrafos HTML
        let formatted = content.replace(/\n\n/g, '</p><p>');
        formatted = formatted.replace(/\n/g, '<br>');
        formatted = '<p>' + formatted + '</p>';

        // Mejorar formato de secciones
        formatted = formatted.replace(/### Parte (\d+)\/(\d+) ###/g, '<h3>Parte $1/$2</h3>');
        formatted = formatted.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>');
        formatted = formatted.replace(/\*(.*?)\*/g, '<em>$1</em>');

        return formatted;
    } catch (error) {
        console.error('Error en formatAnalysisContent:', error);
        return '<p>Error al formatear el contenido del análisis</p>';
    }
}

function downloadPDF() {
    const { jsPDF } = window.jspdf;
    const doc = new jsPDF();

    // Configurar fuente y tamaño
    doc.setFont("helvetica", "normal");
    doc.setFontSize(16);

    // Título
    doc.text("Análisis de Contrato", 20, 30);

    // Fecha
    const fecha = new Date().toLocaleDateString('es-ES');
    doc.setFontSize(12);
    doc.text(`Fecha: ${fecha}`, 20, 45);

    // Archivo analizado
    if (selectedFile) {
        doc.text(`Archivo: ${selectedFile.name}`, 20, 55);
    }

    // Obtener contenido formateado y convertir a texto plano para PDF
    const contentElement = resultsContent;
    const content = contentElement.innerText || contentElement.textContent;

    // Limpiar caracteres especiales y formatear para PDF
    const cleanContent = content.replace(/<[^>]*>/g, ''); // Remover HTML tags
    const lines = doc.splitTextToSize(cleanContent, 170);

    let y = 70;
    doc.setFontSize(10);

    lines.forEach(line => {
        if (y > 270) { // Nueva página si es necesario
            doc.addPage();
            y = 30;
        }
        doc.text(line, 20, y);
        y += 5;
    });

    // Descargar
    doc.save('analisis-reporte-contrato.pdf');
}

// Settings modal functions
function openSettingsModal() {
    loadSettings();
    if (settingsModal) {
        settingsModal.style.display = 'flex';
    } else {
        console.error('Settings modal not found');
    }
}

function closeSettingsModal() {
    if (settingsModal) {
        settingsModal.style.display = 'none';
    }
}

function toggleLLMSettings() {
    if (!llmType || !localSettings || !onlineSettings) {
        console.error('Required DOM elements not found for toggleLLMSettings');
        return;
    }

    if (llmType.value === 'local') {
        localSettings.style.display = 'block';
        onlineSettings.style.display = 'none';
    } else {
        localSettings.style.display = 'none';
        onlineSettings.style.display = 'block';
    }
}

function loadSettings() {
    try {
        const saved = localStorage.getItem('llmConfig');
        if (saved) {
            const parsedConfig = JSON.parse(saved);
            llmConfig = parsedConfig;
        }

        // Cargar valores en los inputs
        if (llmType) llmType.value = llmConfig.type || 'local';
        if (localUrl) localUrl.value = llmConfig.localUrl || '';
        if (apiUrl) apiUrl.value = llmConfig.apiUrl || '';
        if (apiKey) apiKey.value = llmConfig.apiKey || '';

        // Set model name based on type
        const localModelName = document.getElementById('localModelName');
        const onlineModelName = document.getElementById('onlineModelName');
        if (localModelName) localModelName.value = llmConfig.modelName || '';
        if (onlineModelName) onlineModelName.value = llmConfig.modelName || '';

        if (maxTokens) maxTokens.value = llmConfig.maxTokens || 800;

        toggleLLMSettings();
    } catch (error) {
        // En caso de error, usar configuración por defecto
        llmConfig = {
            type: 'local',
            localUrl: 'http://localhost:1234/v1/chat/completions',
            apiUrl: '',
            apiKey: '',
            modelName: '',
            maxTokens: 800
        };
        showError('❌ Error al cargar configuración guardada, usando valores por defecto');
        setTimeout(() => hideError(), 3000);
    }
}

function saveSettings() {
    try {
        // Obtener valores de los inputs
        const localModelName = document.getElementById('localModelName');
        const onlineModelName = document.getElementById('onlineModelName');

        const newConfig = {
            type: llmType.value,
            localUrl: localUrl ? localUrl.value.trim() : '',
            apiUrl: apiUrl ? apiUrl.value.trim() : '',
            apiKey: apiKey ? apiKey.value.trim() : '',
            modelName: (llmType.value === 'local' ?
                (localModelName ? localModelName.value.trim() : '') :
                (onlineModelName ? onlineModelName.value.trim() : '')),
            maxTokens: parseInt(maxTokens ? maxTokens.value : '800') || 800
        };

        // Validar configuración antes de guardar
        let validationError = null;

        if (newConfig.type === 'local') {
            if (!newConfig.localUrl) {
                validationError = '❌ URL del servidor local es requerida';
            } else if (!newConfig.modelName) {
                validationError = '❌ Nombre del Modelo es requerido para modelos locales';
            }
        } else if (newConfig.type === 'online') {
            if (!newConfig.apiUrl) {
                validationError = '❌ URL de API es requerida para modelos en línea';
            } else if (!newConfig.apiKey) {
                validationError = '❌ API Key es requerida para modelos en línea';
            } else if (!newConfig.modelName) {
                validationError = '❌ Nombre del Modelo es requerido para modelos en línea';
            }
        }

        if (validationError) {
            showError(validationError);
            // Fallback: mostrar alert si showError no funciona
            setTimeout(() => {
                if (errorSection.style.display !== 'block') {
                    alert(validationError.replace('❌ ', ''));
                }
            }, 100);
            return;
        }

        // Intentar guardar en localStorage
        localStorage.setItem('llmConfig', JSON.stringify(newConfig));

        // Verificar que se guardó correctamente
        const saved = localStorage.getItem('llmConfig');
        if (!saved) {
            throw new Error('No se pudo guardar en localStorage');
        }

        const parsedSaved = JSON.parse(saved);
        if (JSON.stringify(parsedSaved) !== JSON.stringify(newConfig)) {
            throw new Error('Los datos guardados no coinciden con los enviados');
        }

        // Actualizar configuración global
        llmConfig = newConfig;

        // Actualizar UI
        updateLLMStatusIndicator();
        if (typeof closeSettingsModal === 'function') {
            closeSettingsModal();
        }

        // Mostrar mensaje de éxito
        const modelType = llmConfig.type === 'online' ? 'LLM en línea' : 'LLM local';
        const successMessage = '✓ Configuración guardada: ' + modelType + ' (' + llmConfig.modelName + ')';
        showSuccess(successMessage, 3000);

    } catch (error) {
        const errorMessage = '❌ Error al guardar configuración: ' + error.message;
        showError(errorMessage);

        // Fallback: mostrar alert si showError no funciona
        setTimeout(() => {
            if (errorSection.style.display !== 'block') {
                alert('Error al guardar configuración: ' + error.message);
            }
        }, 100);
    }
}

// Initialize status indicator when DOM is ready
document.addEventListener('DOMContentLoaded', () => {
    // Ensure all DOM elements are available before initializing
    const requiredElements = ['llmIndicator', 'settingsBtn', 'settingsModal'];
    const allElementsPresent = requiredElements.every(id => document.getElementById(id) !== null);

    if (allElementsPresent) {
        updateLLMStatusIndicator();
    } else {
        console.warn('Some required DOM elements are missing, skipping initialization');
    }
});

// Recargar historial al cambiar tamaño de ventana (para responsive)
let resizeTimer;
window.addEventListener('resize', () => {
    clearTimeout(resizeTimer);
    resizeTimer = setTimeout(() => {
        // Solo recargar si el modal de historial está abierto
        if (historyModal && historyModal.style.display === 'flex') {
            const currentContracts = historyList.querySelector('.history-table, .history-cards');
            if (currentContracts) {
                // Obtener los contratos actuales y re-renderizar
                fetch('/api/contracts?limit=50')
                    .then(response => response.json())
                    .then(data => {
                        if (data.success && data.data) {
                            displayHistory(data.data);
                        }
                    })
                    .catch(error => console.error('Error reloading history:', error));
            }
        }
    }, 250);
});

// History functions
function openHistoryModal() {
    historyModal.style.display = 'flex';
    loadHistory();
    loadStats();
}

function closeHistoryModalFn() {
    historyModal.style.display = 'none';
}

function loadHistory() {
    historyList.innerHTML = '<div class="loading-spinner"></div><p>Cargando historial...</p>';
    
    fetch('/api/contracts?limit=50')
        .then(response => response.json())
        .then(data => {
            if (data.success && data.data) {
                displayHistory(data.data);
            } else {
                historyList.innerHTML = '<p>No hay contratos en el historial</p>';
            }
        })
        .catch(error => {
            console.error('Error loading history:', error);
            historyList.innerHTML = '<p class="error">Error al cargar el historial</p>';
        });
}

function searchHistory() {
    const query = historySearch.value.trim();
    
    if (query === '') {
        loadHistory();
        return;
    }
    
    historyList.innerHTML = '<div class="loading-spinner"></div><p>Buscando...</p>';
    
    fetch(`/api/contracts/search?q=${encodeURIComponent(query)}&limit=50`)
        .then(response => response.json())
        .then(data => {
            if (data.success && data.data) {
                displayHistory(data.data);
            } else {
                historyList.innerHTML = '<p>No se encontraron resultados</p>';
            }
        })
        .catch(error => {
            console.error('Error searching:', error);
            historyList.innerHTML = '<p class="error">Error en la búsqueda</p>';
        });
}

function displayHistory(contracts) {
    if (contracts.length === 0) {
        historyList.innerHTML = '<p>No hay contratos en el historial</p>';
        return;
    }
    
    // Detectar si es móvil
    const isMobile = window.innerWidth <= 768;
    
    if (isMobile) {
        // Vista de tarjetas para móviles
        let html = '<div class="history-cards">';
        contracts.forEach(contract => {
            const date = new Date(contract.uploaded_at).toLocaleString('es-ES', {
                day: '2-digit',
                month: '2-digit',
                year: 'numeric',
                hour: '2-digit',
                minute: '2-digit'
            });
            const status = getStatusBadge(contract.status);
            const time = contract.processing_time_seconds ? 
                `${contract.processing_time_seconds.toFixed(1)}s` : '-';
            const llmType = contract.llm_type === 'online' ? '🔵 Online' : '🟢 Local';
            
            html += `
                <div class="history-card">
                    <div class="history-card-header">
                        <strong>📄 ${escapeHtml(contract.filename)}</strong>
                        ${status}
                    </div>
                    <div class="history-card-body">
                        <div class="history-card-row">
                            <span class="label">Fecha:</span>
                            <span>${date}</span>
                        </div>
                        <div class="history-card-row">
                            <span class="label">Modelo:</span>
                            <span>${llmType}</span>
                        </div>
                        <div class="history-card-row">
                            <span class="label">Tiempo:</span>
                            <span>${time}</span>
                        </div>
                    </div>
                    <div class="history-card-actions">
                        ${contract.status === 'completed' ? 
                            `<button class="btn-small" onclick="viewContract(${contract.id})">👁️ Ver</button>` : ''}
                        <button class="btn-small btn-danger" onclick="deleteContract(${contract.id})">🗑️ Eliminar</button>
                    </div>
                </div>
            `;
        });
        html += '</div>';
        historyList.innerHTML = html;
    } else {
        // Vista de tabla para escritorio
        let html = '<table class="history-table"><thead><tr>';
        html += '<th>Archivo</th><th>Fecha</th><th>Estado</th><th>Modelo</th><th>Tiempo</th><th>Acciones</th>';
        html += '</tr></thead><tbody>';
        
        contracts.forEach(contract => {
            const date = new Date(contract.uploaded_at).toLocaleString('es-ES');
            const status = getStatusBadge(contract.status);
            const time = contract.processing_time_seconds ? 
                `${contract.processing_time_seconds.toFixed(1)}s` : '-';
            
            html += `<tr>
                <td><strong>${escapeHtml(contract.filename)}</strong></td>
                <td>${date}</td>
                <td>${status}</td>
                <td>${contract.llm_type === 'online' ? '🔵 Online' : '🟢 Local'}</td>
                <td>${time}</td>
                <td class="history-actions">
                    ${contract.status === 'completed' ? 
                        `<button class="btn-small" onclick="viewContract(${contract.id})">👁️ Ver</button>` : ''}
                    <button class="btn-small btn-danger" onclick="deleteContract(${contract.id})">🗑️</button>
                </td>
            </tr>`;
        });
        
        html += '</tbody></table>';
        historyList.innerHTML = html;
    }
}

function getStatusBadge(status) {
    const badges = {
        'completed': '<span class="badge badge-success">✓ Completado</span>',
        'failed': '<span class="badge badge-error">✗ Fallido</span>',
        'analyzing': '<span class="badge badge-warning">⏳ Analizando</span>',
        'pending': '<span class="badge badge-info">⏸️ Pendiente</span>'
    };
    return badges[status] || status;
}

function loadStats() {
    fetch('/api/contracts/stats')
        .then(response => response.json())
        .then(data => {
            if (data.success && data.data) {
                const stats = data.data;
                document.getElementById('statTotal').textContent = stats.TotalContracts || 0;
                document.getElementById('statCompleted').textContent = stats.CompletedContracts || 0;
                document.getElementById('statFailed').textContent = stats.FailedContracts || 0;
                document.getElementById('statAvgTime').textContent = 
                    stats.AverageProcessingTime ? `${stats.AverageProcessingTime.toFixed(1)}s` : '-';
            }
        })
        .catch(error => {
            console.error('Error loading stats:', error);
        });
}

function viewContract(id) {
    fetch(`/api/contracts/get?id=${id}`)
        .then(response => response.json())
        .then(data => {
            if (data.success && data.data) {
                closeHistoryModalFn();
                // Pasar información del contrato para mostrar en el encabezado
                const contractInfo = {
                    filename: data.data.filename,
                    uploaded_at: data.data.uploaded_at,
                    llm_type: data.data.llm_type,
                    llm_model: data.data.llm_model
                };
                showResults(data.data.analysis_result, contractInfo);
            }
        })
        .catch(error => {
            console.error('Error loading contract:', error);
            alert('Error al cargar el contrato');
        });
}

function deleteContract(id) {
    if (!confirm('¿Estás seguro de que quieres eliminar este contrato del historial?')) {
        return;
    }
    
    fetch(`/api/contracts/delete?id=${id}`, { method: 'POST' })
        .then(response => response.json())
        .then(data => {
            if (data.success) {
                loadHistory();
                loadStats();
            } else {
                alert('Error al eliminar el contrato');
            }
        })
        .catch(error => {
            console.error('Error deleting contract:', error);
            alert('Error al eliminar el contrato');
        });
}

function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}