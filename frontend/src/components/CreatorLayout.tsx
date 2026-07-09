import { Link, Outlet, useLocation, useNavigate } from 'react-router-dom';
import {
  LayoutDashboard, BookOpen, DollarSign, Ticket, Globe, Star, LogOut,
  ChevronDown, Bell,
} from 'lucide-react';
import { useAuth } from '../contexts/AuthContext';
import { useTenant } from '../contexts/TenantContext';
import { useState, useRef, useEffect } from 'react';

const navItems = [
  { path: '/studio/dashboard', label: 'Dashboard', icon: LayoutDashboard },
  { path: '/studio/courses', label: 'Courses', icon: BookOpen },
  { path: '/studio/sales', label: 'Sales', icon: DollarSign },
  { path: '/studio/payouts', label: 'Payouts', icon: DollarSign },
  { path: '/studio/coupons', label: 'Coupons', icon: Ticket },
  { path: '/studio/domain', label: 'Custom Domain', icon: Globe },
  { path: '/studio/reviews', label: 'Reviews', icon: Star },
];

export default function CreatorLayout() {
  const location = useLocation();
  const navigate = useNavigate();
  const { user, logout } = useAuth();
  const { activeTenant } = useTenant();
  const [menuOpen, setMenuOpen] = useState(false);
  const [notifOpen, setNotifOpen] = useState(false);
  const menuRef = useRef<HTMLDivElement>(null);
  const notifRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handler = (e: MouseEvent) => {
      if (menuRef.current && !menuRef.current.contains(e.target as Node)) setMenuOpen(false);
      if (notifRef.current && !notifRef.current.contains(e.target as Node)) setNotifOpen(false);
    };
    document.addEventListener('mousedown', handler);
    return () => document.removeEventListener('mousedown', handler);
  }, []);

  const isActive = (path: string) => location.pathname === path || location.pathname.startsWith(path + '/');

  return (
    <div className="min-h-screen bg-gray-50 flex">
      {/* Sidebar */}
      <aside className="w-64 bg-gray-900 text-gray-300 flex flex-col fixed h-full">
        <div className="p-6 border-b border-gray-800">
          <Link to="/studio/dashboard" className="flex items-center gap-2">
            <BookOpen className="w-6 h-6 text-indigo-400" />
            <span className="font-bold text-white text-lg">{activeTenant?.tenantName || 'Creator Studio'}</span>
          </Link>
        </div>

        <nav className="flex-1 p-4 space-y-1">
          {navItems.map((item) => {
            const Icon = item.icon;
            return (
              <Link
                key={item.path}
                to={item.path}
                className={`flex items-center gap-3 px-3 py-2 rounded-lg text-sm font-medium transition-colors ${
                  isActive(item.path)
                    ? 'bg-indigo-600 text-white'
                    : 'text-gray-400 hover:bg-gray-800 hover:text-white'
                }`}
              >
                <Icon className="w-4 h-4" />
                {item.label}
              </Link>
            );
          })}
        </nav>

        <div className="p-4 border-t border-gray-800">
          <Link
            to="/dashboard"
            className="flex items-center gap-2 text-xs text-gray-500 hover:text-gray-300"
          >
            <LogOut className="w-3 h-3" />
            Back to App
          </Link>
        </div>
      </aside>

      {/* Main content */}
      <div className="flex-1 ml-64">
        {/* Top bar */}
        <header className="bg-white border-b border-gray-200 px-6 py-3 flex items-center justify-between sticky top-0 z-10">
          <h2 className="text-lg font-semibold text-gray-800">
            {navItems.find(n => isActive(n.path))?.label || 'Creator Studio'}
          </h2>

          <div className="flex items-center gap-4">
            {/* Notifications */}
            <div ref={notifRef} className="relative">
              <button
                onClick={() => setNotifOpen(!notifOpen)}
                className="p-2 rounded-lg hover:bg-gray-100 relative"
              >
                <Bell className="w-5 h-5 text-gray-600" />
                <span className="absolute top-1 right-1 w-2 h-2 bg-red-500 rounded-full"></span>
              </button>
              {notifOpen && (
                <div className="absolute right-0 mt-2 w-80 bg-white rounded-lg shadow-lg border border-gray-200 py-2">
                  <div className="px-4 py-2 text-sm text-gray-500">No new notifications</div>
                </div>
              )}
            </div>

            {/* User menu */}
            <div ref={menuRef} className="relative">
              <button
                onClick={() => setMenuOpen(!menuOpen)}
                className="flex items-center gap-2 px-3 py-1.5 rounded-lg hover:bg-gray-100"
              >
                <div className="w-8 h-8 bg-indigo-500 rounded-full flex items-center justify-center text-white text-sm font-medium">
                  {user?.displayName?.[0]?.toUpperCase() || 'U'}
                </div>
                <span className="text-sm font-medium text-gray-700">{user?.displayName}</span>
                <ChevronDown className="w-4 h-4 text-gray-400" />
              </button>
              {menuOpen && (
                <div className="absolute right-0 mt-2 w-48 bg-white rounded-lg shadow-lg border border-gray-200 py-2">
                  <Link to="/settings" className="block px-4 py-2 text-sm text-gray-700 hover:bg-gray-50">
                    Settings
                  </Link>
                  <button
                    onClick={() => { logout(); navigate('/login'); }}
                    className="block w-full text-left px-4 py-2 text-sm text-gray-700 hover:bg-gray-50"
                  >
                    Sign out
                  </button>
                </div>
              )}
            </div>
          </div>
        </header>

        {/* Page content */}
        <main className="p-6">
          <Outlet />
        </main>
      </div>
    </div>
  );
}
