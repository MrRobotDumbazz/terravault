import { Routes, Route } from "react-router-dom";
import Navbar from "@/components/ui/Navbar";
import { useAuth } from "@/hooks/useAuth";
import Landing from "@/pages/Landing";
import ProjectListing from "@/pages/ProjectListing";
import ProjectDetail from "@/pages/ProjectDetail";
import InvestorDashboard from "@/pages/InvestorDashboard";
import DeveloperDashboard from "@/pages/DeveloperDashboard";
import DeveloperProjectManage from "@/pages/DeveloperProjectManage";
import KYCFlow from "@/pages/KYCFlow";
import AdminPanel from "@/pages/AdminPanel";

export default function App() {
  useAuth();
  return (
    <div className="min-h-screen bg-gray-950">
      <Navbar />
      <main>
        <Routes>
          <Route path="/" element={<Landing />} />
          <Route path="/projects" element={<ProjectListing />} />
          <Route path="/projects/:id" element={<ProjectDetail />} />
          <Route path="/dashboard" element={<InvestorDashboard />} />
          <Route path="/developer" element={<DeveloperDashboard />} />
          <Route path="/developer/projects/:id" element={<DeveloperProjectManage />} />
          <Route path="/kyc" element={<KYCFlow />} />
          <Route path="/admin" element={<AdminPanel />} />
          <Route path="*" element={
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10 flex flex-col items-center justify-center min-h-[60vh]">
              <div className="text-6xl font-bold text-gray-700 mb-4">404</div>
              <h1 className="text-2xl font-semibold text-white mb-2">Page Not Found</h1>
              <p className="text-gray-400">The page you're looking for doesn't exist.</p>
            </div>
          } />
        </Routes>
      </main>
    </div>
  );
}
