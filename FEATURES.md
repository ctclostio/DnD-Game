# D&D Game Features Documentation

## API Documentation (OpenAPI/Swagger)

### Overview
The application provides comprehensive API documentation using OpenAPI/Swagger specification.

### Access
- **Swagger UI**: Navigate to `/swagger` to view interactive API documentation
- **OpenAPI JSON**: Available at `/api/v1/swagger.json`

### Features
- Interactive API testing interface
- Full endpoint documentation with request/response schemas
- Authentication support with JWT bearer tokens
- Request/response examples for all endpoints

### Implementation
- Swagger annotations on all handler functions
- Automatic documentation generation
- Versioned API with clear schemas

## Health Check Endpoints

### Available Endpoints

#### Basic Health Check
- **Endpoint**: `GET /health`
- **Description**: Quick health status check
- **Response**: Basic health status with timestamp and version

#### Liveness Probe
- **Endpoint**: `GET /health/live`
- **Description**: Kubernetes liveness probe endpoint
- **Purpose**: Determines if the service should be restarted
- **Response**: Simple alive status

#### Readiness Probe
- **Endpoint**: `GET /health/ready`
- **Description**: Kubernetes readiness probe endpoint
- **Purpose**: Determines if the service is ready to accept traffic
- **Checks**:
  - Database connectivity
  - Required services initialization
  - WebSocket hub status

#### Detailed Health Check
- **Endpoint**: `GET /health/detailed` (Requires authentication)
- **Description**: Comprehensive health information
- **Includes**:
  - System metrics (CPU, memory, goroutines)
  - Component health status
  - Uptime information
  - Go runtime statistics

## Internationalization (i18n)

### Supported Languages
- English (en) - Default
- Spanish (es)

### Features

#### Translation System
- Centralized translation management
- Nested translation keys for organization
- Parameter interpolation support
- Fallback to English for missing translations

#### Locale Management
- Automatic locale detection from browser
- Persistent locale selection (localStorage)
- Easy language switching via UI

#### Formatting
- Number formatting based on locale
- Date/time formatting with locale support
- Relative time formatting ("2 hours ago")

#### Usage
```typescript
import { useTranslation } from '../hooks/useTranslation';

const { t, locale, setLocale } = useTranslation();
const greeting = t('common.hello', { name: 'Player' });
```

### Adding New Languages
1. Create translation file in `frontend/src/i18n/translations/`
2. Import and add to translations object in `frontend/src/i18n/index.ts`
3. Update language selector options

## Accessibility Features

### Screen Reader Support
- Semantic HTML throughout the application
- ARIA labels and descriptions for interactive elements
- Live regions for dynamic content updates
- Screen reader announcements for important actions

### Keyboard Navigation
- Full keyboard accessibility for all interactive elements
- Skip links to main content
- Focus trap for modals and dialogs
- Keyboard shortcuts:
  - `1` - Skip to main content
  - `F6` - Navigate between landmarks
  - `Tab/Shift+Tab` - Navigate interactive elements
  - `Escape` - Close modals and menus

### Visual Accessibility

#### High Contrast Mode
- Increased color contrast for better visibility
- Clear borders and outlines
- Enhanced text readability

#### Large Text Mode
- 120% base font size
- Proportionally scaled headings
- Maintained layout integrity

#### Reduced Motion
- Respects `prefers-reduced-motion` system preference
- Disables animations and transitions
- Instant state changes for better predictability

#### Enhanced Focus Indicators
- Visible focus outlines on all interactive elements
- High contrast focus rings
- Configurable focus indicator styles

### Accessibility Settings
- User-configurable accessibility preferences
- Settings persist across sessions
- Easy toggle interface for all options

### Component Features

#### Skip Links
- Skip to main content link
- Visible on focus
- Improves navigation efficiency

#### Focus Management
- Automatic focus management for route changes
- Return focus after modal/dialog close
- Focus trap utility for contained interactions

#### Announcements
- Live region announcements for dynamic updates
- Error and success message announcements
- Loading state announcements

#### Modal Accessibility
- Proper ARIA attributes
- Focus trap implementation
- Escape key handling
- Background scroll prevention

### Developer Tools

#### useAnnounce Hook
```typescript
const { announce, announceError, announceSuccess } = useAnnounce();
announceSuccess('Character saved successfully');
```

#### useFocusTrap Hook
```typescript
const modalRef = useFocusTrap({
  enabled: isOpen,
  returnFocus: true,
});
```

#### AccessibilityProvider
- Global accessibility settings management
- System preference detection
- CSS class application for visual modes

### Testing Accessibility
1. Test with screen readers (NVDA, JAWS, VoiceOver)
2. Navigate using only keyboard
3. Check color contrast ratios
4. Validate ARIA attributes
5. Test with browser accessibility tools

### Compliance
- WCAG 2.1 Level AA compliance target
- Semantic HTML5 elements
- Proper heading hierarchy
- Alternative text for images
- Form label associations