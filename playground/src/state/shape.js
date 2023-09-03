import {create} from 'zustand';
import axios from 'axios'; // Assuming you're using axios as the "client"
import yaml from 'js-yaml';
import {toast} from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';

export const useShapeStore = create((set, get) => ({
    // main shape items
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
    systemError: "",
    schemaError: null,
    relationshipErrors: [],
    attributeErrors: [],
    visualizerError: "",
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

    removeScenario: (index) => {
        set((state) => ({
            scenarios: state.scenarios.filter((_, idx) => idx !== index)
        }));
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
                set({ runLoading: false });
            })
        }, 500);
    },

    run: () => {
        get().clearErrors()

        const shape = {
            schema: get().schema,
            relationships: get().relationships,
            attributes: get().attributes,
            scenarios: get().scenarios,
        }

        Run(JSON.stringify(shape, null, 2))
            .then((rr) => {
                for (let i = 0; i < rr.length; i++) {
                    const error = JSON.parse(rr[i]);

                    if (error.type === 'file_validation') {
                        set({systemError: error.message});
                        toast.error(`System Error: ${error.message}`);
                    }

                    if (error.type === 'schema') {
                        set({schemaError: handleSchemaError(error.message)});
                    }

                    if (error.type === 'relationships') {
                        set((state) => ({relationshipErrors: [...state.relationshipErrors, error]}));
                    }

                    if (error.type === 'attributes') {
                        set((state) => ({attributeErrors: [...state.attributeErrors, error]}));
                    }

                    if (error.type === 'scenarios') {
                        set((state) => ({scenariosError: [...state.scenariosError, error]}));
                    }
                }

                // If there were no schema errors, proceed to visualize
                if (!rr.some(error => JSON.parse(error).type === 'schema') && !rr.some(error => JSON.parse(error).type === 'file_validation')) {
                    return Visualize();
                } else {
                    // You can add additional error handling or set state as needed
                    set({
                        graph: {nodes: [], edges: []},
                    });
                }
            }).then((vr) => {
            if (vr) {  // Only execute if vr exists (means no errors previously)
                set({
                    graph: JSON.parse(vr[0]),
                    definitions: JSON.parse(vr[1]),
                });
            }
        }).catch((error) => {
            toast.error(`System Error: ${error.message}`);
            set({systemError: error})
        });
    },

    // Clear the state
    clear: () => set({
        schema: ``,
        relationships: [],
        attributes: [],
        graph: {nodes: [], edges: []},
        systemError: "",
        schemaError: null,
        visualizerError: "",
        relationshipErrors: [],
        attributeErrors: [],
        scenariosError: ""
    }),

    // Clear the state
    clearErrors: () => set({
        systemError: "",
        schemaError: null,
        visualizerError: "",
        relationshipErrors: [],
        attributeErrors: [],
        scenariosError: ""
    }),

    fetchShape: async (pond) => {
        try {
            get().clear()

            const response = await axios.get(`https://s3.amazonaws.com/permify.playground.storage/shapes/${pond}.yaml`);
            const result = yaml.load(response.data, null);

            get().setSchema(result.schema ?? ``);
            get().addRelationships(result.relationships ?? []);
            get().addAttributes(result.attributes ?? []);
            get().setScenarios(result.scenarios ?? []);

        } catch (error) {
            toast.error(`System Error: ${error.message}`);
            set({systemError: error})
        }
    },

    getEntityTypes: () => {
        const entityDefinitions = get()?.definitions?.entityDefinitions;
        return Object.keys(entityDefinitions || {});
    },

    getRelations: (entityType) => {
        const relations = get()?.definitions?.entityDefinitions?.[entityType]?.relations;
        return Object.keys(relations || {});
    },

    getSubjectTypes: (entityType, relation) => {
        const references = get()?.definitions?.entityDefinitions?.[entityType]?.relations?.[relation]?.relationReferences;

        if (!references || references.length === 0) {
            return [];
        }

        const uniqueTypes = new Set(references.map(ref => ref?.type).filter(Boolean));
        return [...uniqueTypes];
    },

    getSubjectRelations: (entityType, relation, subjectType) => {
        const references = get()?.definitions?.entityDefinitions?.[entityType]?.relations?.[relation]?.relationReferences;

        if (!references || references.length === 0) {
            return [];
        }

        return references
            .filter(v => v?.type === subjectType)
            .map(v => v?.relation)
            .filter(Boolean);
    },

    getAttributes: (entityType) => {
        const attributes = get()?.definitions?.entityDefinitions?.[entityType]?.attributes;
        return Object.keys(attributes || {});
    },

    getTypeValueBasedOnAttribute: (entityType, attribute) => {
        const type = get()?.definitions?.entityDefinitions?.[entityType]?.attributes[attribute]?.type;
        switch (type) {
            case "ATTRIBUTE_TYPE_BOOLEAN":
                return "boolean"
            case "ATTRIBUTE_TYPE_BOOLEAN_ARRAY":
                return "boolean][]"
            case "ATTRIBUTE_TYPE_STRING":
                return "string"
            case "ATTRIBUTE_TYPE_STRING_ARRAY":
                return "string[]"
            case "ATTRIBUTE_TYPE_INTEGER":
                return "integer"
            case "ATTRIBUTE_TYPE_INTEGER_ARRAY":
                return "integer[]"
            case "ATTRIBUTE_TYPE_DOUBLE":
                return "double"
            case "ATTRIBUTE_TYPE_DOUBLE_ARRAY":
                return "double[]"
            default:
                return ""
        }
    },
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
