# SpaceNet Project Configuration

## Project Overview

SpaceNet is an innovative space-themed IPv6 network visualization and claiming system that combines sophisticated backend networking with immersive 3D frontend experiences. The project features a distributed architecture with Go-based backend services and a Next.js frontend with Three.js 3D visualizations.

### Technology Stack Analysis

**Backend Architecture:**
- **Language**: Go 1.24.5
- **Web Framework**: Gorilla Mux for HTTP routing
- **Database**: Redis for distributed claim storage
- **Networking**: HTTP API for IPv6 claims with proof-of-work validation
- **Architecture**: HTTP-based microservices with RESTful API endpoints

**Frontend Architecture:**
- **Framework**: Next.js 15.4.4 with React 19
- **3D Engine**: Three.js with React Three Fiber
- **UI Library**: Tailwind CSS 4.0
- **Language**: TypeScript 5
- **Visualization**: Custom 3D space components (galaxies, star clusters, solar systems)

**Infrastructure:**
- **Containerization**: Docker Compose
- **Caching**: Redis with persistence
- **Networking**: IPv6-first design

## AI Development Team Configuration

### Core Development Specialists

#### **Backend Development** → @backend-developer
- **Primary Focus**: Go backend services, networking protocols, distributed systems
- **Expertise**: HTTP server optimization, Redis integration, concurrent processing
- **Use Cases**: 
  - "Add new API endpoints for claim statistics"
  - "Implement distributed claim validation logic"
  - "Optimize HTTP request processing and proof-of-work validation"

#### **Frontend Development** → @react-nextjs-expert
- **Primary Focus**: Next.js application architecture, React patterns, TypeScript optimization
- **Expertise**: Modern React hooks, state management, Next.js 15 features, performance optimization
- **Use Cases**:
  - "Refactor the SpaceNet browser component structure"
  - "Add real-time updates to the subnet table"
  - "Implement client-side caching for network claims"

#### **3D Visualization Specialist** → @frontend-developer
- **Primary Focus**: Three.js integration, React Three Fiber, 3D scene optimization
- **Expertise**: WebGL shaders, 3D animations, spatial data visualization, performance tuning
- **Use Cases**:
  - "Create new space objects for deeper network levels"
  - "Optimize shader performance for galaxy rendering"
  - "Add interactive animations for network traversal"

### Specialized System Experts

#### **API Architecture** → @api-architect
- **Primary Focus**: RESTful API design, HTTP protocol optimization, distributed system patterns
- **Expertise**: API versioning, rate limiting, proof-of-work integration
- **Use Cases**:
  - "Design scalable endpoints for massive IPv6 address spaces"
  - "Optimize claim submission API for high throughput"
  - "Add authentication and authorization layers"

#### **Performance Optimization** → @performance-optimizer
- **Primary Focus**: System-wide performance analysis, bottleneck identification, scaling solutions
- **Expertise**: Go profiling, Redis optimization, React rendering performance, WebGL optimization
- **Use Cases**:
  - "Profile and optimize the claim processing pipeline"
  - "Reduce memory usage in 3D scene rendering"
  - "Analyze and improve subnet generation performance"

#### **UI/UX Enhancement** → @tailwind-css-expert
- **Primary Focus**: Modern UI design, responsive layouts, accessibility, design systems
- **Expertise**: Tailwind CSS 4.0, component design, dark mode, responsive 3D layouts
- **Use Cases**:
  - "Create a cohesive space-themed design system"
  - "Improve mobile responsiveness for 3D views"
  - "Add accessibility features for network navigation"

### Quality Assurance Specialists

#### **Code Review** → @code-reviewer
- **Primary Focus**: Code quality, security analysis, best practices enforcement
- **Expertise**: Go code review, React/TypeScript analysis, distributed system patterns
- **Use Cases**:
  - "Review my claim processing security implementation"
  - "Analyze the 3D component architecture for maintainability"
  - "Security audit of the HTTP API implementation"

#### **Documentation** → @documentation-specialist
- **Primary Focus**: Technical documentation, API documentation, architectural guides
- **Expertise**: Go documentation, TypeScript interfaces, system architecture diagrams
- **Use Cases**:
  - "Document the IPv6 claiming protocol"
  - "Create API reference for the HTTP endpoints"
  - "Write deployment guides for the Docker setup"

## Team Usage Guide

### Quick Commands for Common Tasks

**Backend Development:**
- "Add rate limiting to the claim API"
- "Optimize Redis connection pooling"
- "Implement claim conflict resolution"

**Frontend Development:**
- "Add smooth transitions between network levels"
- "Create a real-time claim status indicator"
- "Implement keyboard shortcuts for navigation"

**3D Visualization:**
- "Add particle effects for active claims"
- "Create dynamic camera paths for network exploration"
- "Implement level-of-detail for performance scaling"

**System Architecture:**
- "Design a microservices deployment strategy"
- "Add monitoring and observability features"
- "Implement horizontal scaling for claim processing"

### Project-Specific Workflows

#### Adding New Network Levels
1. Backend: Extend the IPv6 address generation logic
2. Frontend: Create new 3D space object components
3. Integration: Update the level transition animations
4. Testing: Verify claim routing at new depths

#### Performance Optimization Cycle
1. Profile: Use Go pprof and React DevTools
2. Identify: Bottlenecks in claim processing or rendering
3. Optimize: Backend concurrency or frontend memoization
4. Validate: Load testing and 3D performance metrics

#### Feature Integration Process  
1. API Design: Define new endpoints and data structures
2. Backend Implementation: Go services and Redis schema
3. Frontend Integration: React components and 3D objects
4. End-to-End Testing: Network claims through 3D interface

## Project Architecture Insights

Your SpaceNet project showcases several advanced patterns:

- **HTTP-First Design**: RESTful API for all operations with proof-of-work validation
- **Hierarchical Visualization**: 8-level IPv6 space mapped to cosmic structures
- **Real-Time Coordination**: Redis for distributed claim state management  
- **Immersive UX**: 3D navigation through abstract network concepts
- **Scalable Processing**: Worker pools for concurrent claim handling

This configuration ensures your AI development team understands both the technical implementation details and the conceptual elegance of mapping network topologies to space exploration metaphors.

---

*Team configuration optimized for SpaceNet's unique blend of systems programming, 3D visualization, and distributed networking. Your AI specialists are ready to help build the universe of IPv6 space exploration.*

🚀 **Ready to explore the network cosmos? Try commands like:**
- "Add stellar animations when claims are processed"
- "Optimize the galaxy rendering for thousands of subnets"  
- "Implement real-time multiplayer claim visualization"