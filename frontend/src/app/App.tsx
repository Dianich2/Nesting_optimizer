import { NestingPage } from '../pages/NestingPage'

import {
    BrowserRouter,
    Navigate,
    Route,
    Routes,
} from 'react-router'

import { AppLayout } from '../components/layout/AppLayout'
import { EditorPage } from '../pages/EditorPage'
import { LibraryPage } from '../pages/LibraryPage'
import { LoginPage } from '../pages/LoginPage'
import { ProjectsPage } from '../pages/ProjectsPage'
import { RegisterPage } from '../pages/RegisterPage'
import { ProtectedRoute } from '../components/auth/ProtectedRoute'


export function App() {
    return (
        <BrowserRouter>
            <Routes>
                <Route path="/login" element={<LoginPage />} />
                <Route path="/register" element={<RegisterPage />} />

                <Route element={<ProtectedRoute />}>
                    <Route element={<AppLayout />}>
                        <Route path="/projects" element={<ProjectsPage />} />
                        <Route path="/projects/:projectId/surfaces/:surfaceId/nesting" element={<NestingPage />} />
                        <Route path="/library" element={<LibraryPage />} />
                        <Route
                            path="/projects/:projectId/editor"
                            element={<EditorPage />}
                        />
                    </Route>
                </Route>

                <Route path="/" element={<Navigate to="/projects" replace />} />
                <Route path="*" element={<Navigate to="/projects" replace />} />
            </Routes>
        </BrowserRouter>
    )
}
