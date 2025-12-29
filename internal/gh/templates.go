package gh

// Template names
const (
	TemplateEpic      = "epic"
	TemplateUserStory = "user-story"
	TemplateTask      = "task"
)

var templates = map[string]string{
	TemplateEpic:      epicTemplate,
	TemplateUserStory: userStoryTemplate,
	TemplateTask:      taskTemplate,
}

// GetTemplate returns the template content for the given template name
func GetTemplate(name string) string {
	if tmpl, ok := templates[name]; ok {
		return tmpl
	}
	return ""
}

// ListTemplates returns all available template names
func ListTemplates() []string {
	names := make([]string, 0, len(templates))
	for name := range templates {
		names = append(names, name)
	}
	return names
}

const epicTemplate = `
### 📝 Descripción

Sistema completo de... incluyendo:
- Funcionalidad 1
- Funcionalidad 2
- Funcionalidad 3


### 🎯 Objetivo

Permitir a los usuarios...

### 📦 Historias Incluidas

### Historia 1: [Nombre]
Descripción breve de la historia

### Historia 2: [Nombre]
Descripción breve de la historia


### 🎯 Acceptance Criteria

- ✅ Criterio 1
- ✅ Criterio 2
- ✅ Criterio 3


### 👥 Equipos Involucrados

- **Backend**: Descripción del trabajo
- **App**: Descripción del trabajo
- **Web**: Descripción del trabajo
- **Auth**: Descripción del trabajo


### 📊 Estimación

- **Historias:** X
- **Tareas estimadas:** ~X
- **Complejidad:** Alta/Media/Baja


### 📝 Notas Técnicas

- Nota técnica 1
- Nota técnica 2
`

const userStoryTemplate = `
### 📝 Historia de Usuario

Como [tipo de usuario]
Quiero [acción/funcionalidad]
Para [beneficio/objetivo]

### 🎯 Acceptance Criteria

- ✅ Criterio 1
- ✅ Criterio 2
- ✅ Criterio 3

### 📋 Tareas

<!-- Las tareas se agregarán automáticamente aquí -->

### 📝 Notas Técnicas

- Nota 1
- Nota 2
`

const taskTemplate = `
### 📝 Descripción

Descripción detallada de la tarea...

### ✅ Checklist

- [ ] Subtarea 1
- [ ] Subtarea 2
- [ ] Subtarea 3

### 📝 Notas

- Nota 1
- Nota 2
`
