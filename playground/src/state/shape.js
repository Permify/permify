import {create} from 'zustand';
import axios from 'axios'; // Assuming you're using axios as the "client"
import yaml from 'js-yaml';

export const useShapeStore = create((set, get) => ({
    schema: ``,
    relationships: [],
    attributes: [],
    scenarios: [],

    graph: {nodes: [], edges: []},
    runLoading: false,

    assertionCount: 0,

    // definitions
    definitions: {},

    // errors
    schemaError: null,
    visualizerError: "",
    relationshipErrors: [],
    attributeErrors: [],
    yamlValidationErrors: [],
    scenariosError: [],

    setSchema: (schema) => {
        set({schema: schema})

        setTimeout(() => {
            get().run()
        }, 500);
    },

    addRelationships: (relationships) => {
        set((state) => ({relationships: [...state.relationships, ...relationships]}));

        setTimeout(() => {
            get().run()
        }, 500);
    },

    setScenarios: (scenarios) => {
        set({scenarios: scenarios});
    },

    setRelationships: (relationships) => {
        set({relationships: relationships});
    },

    setAttributes: (attributes) => {
        set({attributes: attributes});
    },

    removeRelationship: (relationship) => {
        set((state) => ({
            relationships: state.relationships.filter(rid => rid !== relationship)
        }));

        setTimeout(() => {
            get().run()
        }, 500);
    },

    addAttributes: (attributes) => {
        set((state) => ({attributes: [...state.attributes, ...attributes]}));

        setTimeout(() => {
            get().run()
        }, 500);
    },

    removeAttribute: (attribute) => {
        set((state) => ({
            attributes: state.attributes.filter(aid => aid !== attribute)
        }));

        setTimeout(() => {
            get().run()
        }, 500);
    },

    runAsync: async () => {
        try {
            await get().run();
        } catch (error) {
            console.error(error);
        }
    },

    runAssertions: () => {
        set((state) => ({assertionCount: state.assertionCount + 1}))
        set({runLoading: true});
        setTimeout(() => {
            get().runAsync().then(() => {
                set({runLoading: false});
            })
        }, 500);
    },

    run: () => {
        // Clear existing errors
        get().clearAllErrors();

        // Define the shape object
        const shape = {
            schema: get().schema,
            relationships: get().relationships,
            attributes: get().attributes,
            scenarios: get().scenarios,
        };

        // Function to handle errors
        const handleError = (error) => {
            switch (error.type) {
                case 'file_validation':
                    set((state) => ({yamlValidationErrors: [...state.yamlValidationErrors, error]}));
                    break;
                case 'schema':
                    set({schemaError: handleSchemaError(error.message)});
                    break;
                case 'relationships':
                    set((state) => ({relationshipErrors: [...state.relationshipErrors, error]}));
                    break;
                case 'attributes':
                    set((state) => ({attributeErrors: [...state.attributeErrors, error]}));
                    break;
                case 'scenarios':
                    set((state) => ({scenariosError: [...state.scenariosError, error]}));
                    break;
                default:
                    break;
            }
        };

        // Run the shape validation
        Run(JSON.stringify(shape, null, 2))
            .then((response) => {
                // Process each error in the response
                response.forEach((errorString) => {
                    const error = JSON.parse(errorString);
                    handleError(error);
                });

                // If no schema or file validation errors, proceed to visualize
                const hasSchemaError = response.some(error => JSON.parse(error).type === 'schema');
                const hasFileValidationError = response.some(error => JSON.parse(error).type === 'file_validation');

                if (!hasSchemaError && !hasFileValidationError) {
                    return Visualize();
                } else {
                    // Set graph to empty if there are schema or file validation errors
                    set({graph: {nodes: [], edges: []}});
                }
            })
            .then((visualizationResponse) => {
                if (visualizationResponse) {
                    // Update the graph and definitions state if visualization was successful
                    set({
                        graph: JSON.parse(visualizationResponse[0]),
                        definitions: JSON.parse(visualizationResponse[1]),
                    });
                }
            })
            .catch((error) => {
               console.error(error)
            });
    },

    // Clear the state
    clear: () => set({
        schema: ``,
        relationships: [],
        attributes: [],
        graph: {nodes: [], edges: []},
        errors: [],
        schemaError: null,
        visualizerError: "",
        relationshipErrors: [],
        attributeErrors: [],
        scenariosError: ""
    }),

    // Clear the state
    clearAllErrors: () => set({
        schemaError: null,
        visualizerError: "",
        relationshipErrors: [],
        attributeErrors: [],
        scenariosError: [],
    }),

    fetchShape: async (pond) => {
        try {
            get().clear()
            const response = await axios.get(`https://697comsaveyxrwi8.public.blob.vercel-storage.com/${pond}.yaml`);
            const result = yaml.load(response.data, null);

            get().setSchema(result.schema ?? ``);
            get().setRelationships(result.relationships ?? []);
            get().setAttributes(result.attributes ?? []);
            get().setScenarios(result.scenarios ?? []);

        } catch (error) {
            const baseUrl = window.location.href.split('?')[0];
            window.location = `${baseUrl}404`;
        }
    },

    getEntityTypes: () => {
        const entityDefinitions = get().definitions?.entityDefinitions;
        return Object.keys(entityDefinitions || {});
    },

    getRelations: (entityType) => {
        const relations = get().definitions?.entityDefinitions?.[entityType]?.relations;
        return Object.keys(relations || {});
    },

    getSubjectTypes: (entityType, relation) => {
        const references = get().definitions?.entityDefinitions?.[entityType]?.relations?.[relation]?.relationReferences;

        if (!references || references.length === 0) {
            return [];
        }

        const uniqueTypes = new Set(references.map(ref => ref?.type).filter(Boolean));
        return [...uniqueTypes];
    },

    getSubjectRelations: (entityType, relation, subjectType) => {
        const references = get().definitions?.entityDefinitions?.[entityType]?.relations?.[relation]?.relationReferences;

        if (!references || references.length === 0) {
            return [];
        }

        return references
            .filter(v => v?.type === subjectType)
            .map(v => v?.relation)
            .filter(Boolean);
    },

    getAttributes: (entityType) => {
        const attributes = get().definitions?.entityDefinitions?.[entityType]?.attributes;
        return Object.keys(attributes || {});
    },

    getTypeValueBasedOnAttribute: (entityType, attribute) => {
        const type = get().definitions?.entityDefinitions?.[entityType]?.attributes?.[attribute]?.type;
        const typeMap = {
            "ATTRIBUTE_TYPE_BOOLEAN": "boolean",
            "ATTRIBUTE_TYPE_BOOLEAN_ARRAY": "boolean[]",
            "ATTRIBUTE_TYPE_STRING": "string",
            "ATTRIBUTE_TYPE_STRING_ARRAY": "string[]",
            "ATTRIBUTE_TYPE_INTEGER": "integer",
            "ATTRIBUTE_TYPE_INTEGER_ARRAY": "integer[]",
            "ATTRIBUTE_TYPE_DOUBLE": "double",
            "ATTRIBUTE_TYPE_DOUBLE_ARRAY": "double[]"
        };
        return typeMap[type] || "";
    },

    setDefinitions: (definitions) => set({definitions}),

    // ERRORS

    setError: (errorType, error) => set(state => {
        switch (errorType) {
            case 'yamlValidationErrors':
                return {yamlValidationErrors: [...state.yamlValidationErrors, error]};
            case 'schemaError':
                return {schemaError: error};
            case 'relationshipErrors':
                return {relationshipErrors: [...state.relationshipErrors, error]};
            case 'attributeErrors':
                return {attributeErrors: [...state.attributeErrors, error]};
            case 'visualizerError':
                return {visualizerError: error};
            case 'scenariosError':
                return {scenariosError: [...state.scenariosError, error]};
            default:
                return {};
        }
    }),

    clearError: (errorType) => set(state => {
        switch (errorType) {
            case 'yamlValidationErrors':
                return {yamlValidationErrors: []};
            case 'schemaError':
                return {schemaError: null};
            case 'relationshipErrors':
                return {relationshipErrors: []};
            case 'attributeErrors':
                return {attributeErrors: []};
            case 'visualizerError':
                return {visualizerError: ""};
            case 'scenariosError':
                return {scenariosError: []};
            default:
                return {};
        }
    }),
}));

// Handles schema errors.
const handleSchemaError = (res) => {
    // Get line and column numbers from the error message.
    let numbers = parseNumbers(res);

    return {
        schemaError: {
            line: numbers[0],       // Error's line number.
            column: numbers[1],     // Error's column number.
            // Convert the message to a more readable format.
            message: res.replaceAll('_', ' ').toLowerCase()
        },
    };
};

// Gets two numbers from the start of a string.
const parseNumbers = (input) => {
    // Looks for two numbers separated by a colon.
    const regex = /^(\d+):(\d+)/;
    const match = regex.exec(input);

    if (match) {
        // If found, return the numbers.
        return [parseInt(match[1], 10), parseInt(match[2], 10)]
    } else {
        // If not found, return [0, 0].
        return [0, 0]
    }
}

function Run(shape) {
    return new Promise((resolve) => {
        let res = window.run(shape)
        resolve(res);
    });
}

function Visualize() {
    return new Promise((resolve) => {
        let res = window.visualize()
        resolve(res);
    });
}
