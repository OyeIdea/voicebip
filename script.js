// JavaScript for RevoVoiceAI Data Flow Visualizer

document.addEventListener('DOMContentLoaded', () => {
    console.log('RevoVoiceAI Data Flow Visualizer script loaded.');

    // 2. Define Component Sequences
    const flowSequences = {
        realtime: [
            'call-initiation', 'voice-gateway-1', 'real-time-processing-engine',
            'ai-ml-services', 'business-logic-layer', 'text-to-speech',
            'voice-gateway-2', 'customer'
        ],
        agent: [
            'call-context', 'real-time-agent-assistant', 'agent-dashboard',
            'agent-interface', 'agent-actions', 'crm-integration', 'data-management-layer'
        ],
        analytics: [
            'call-events', 'stream-processing', 'monitoring-dashboard',
            'data-warehouse', 'ml-training', 'business-intelligence'
        ]
    };

    // 3. Get Component Elements
    const components = document.querySelectorAll('.component');
    const componentElements = {};
    components.forEach(comp => {
        componentElements[comp.id] = comp;
    });

    // 4. Initialize Flow State
    const flowStates = {
        realtime: { currentIndex: -1, currentActiveElement: null },
        agent: { currentIndex: -1, currentActiveElement: null },
        analytics: { currentIndex: -1, currentActiveElement: null }
    };

    // 5. Click Event Handler Function
    function handleComponentClick(flowName, clickedComponentId) {
        const flowArray = flowSequences[flowName];
        const state = flowStates[flowName];
        const clickedComponentElement = componentElements[clickedComponentId];

        if (!flowArray || !state || !clickedComponentElement) {
            console.error('Error in component click handling:', flowName, clickedComponentId);
            return;
        }

        const clickedIndex = flowArray.indexOf(clickedComponentId);

        // Check if the click is on the next component in sequence or the first one
        if (clickedIndex === state.currentIndex + 1) {
            // Remove 'active' from previously active component in this flow
            if (state.currentActiveElement) {
                state.currentActiveElement.classList.remove('active');
            }

            // Add 'active' to the clicked component
            clickedComponentElement.classList.add('active');
            state.currentActiveElement = clickedComponentElement;
            state.currentIndex = clickedIndex;
        } else if (clickedIndex === 0 && state.currentIndex !== 0) {
            // Allow restarting the flow by clicking the first element
            if (state.currentActiveElement) {
                state.currentActiveElement.classList.remove('active');
            }
            clickedComponentElement.classList.add('active');
            state.currentActiveElement = clickedComponentElement;
            state.currentIndex = 0;
        }
        // Optional: else, if clicked out of order and not the first element, do nothing or reset
        // For now, strict progression or restart by clicking the first item.
    }

    // 6. Attach Event Listeners
    components.forEach(component => {
        const flowName = component.dataset.flow;
        if (flowName) {
            component.addEventListener('click', () => {
                handleComponentClick(flowName, component.id);
            });
        }
    });

    // 7. Initial State: No components active by default. User must click to start.
    // Optional: To activate the first component of each flow on load:
    // Object.keys(flowSequences).forEach(flowName => {
    //     const firstComponentId = flowSequences[flowName][0];
    //     if (componentElements[firstComponentId]) {
    //         componentElements[firstComponentId].classList.add('active');
    //         flowStates[flowName].currentActiveElement = componentElements[firstComponentId];
    //         flowStates[flowName].currentIndex = 0;
    //     }
    // });
});
