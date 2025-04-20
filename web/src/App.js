import { BrowserRouter as Router, Routes, Route } from "react-router-dom";
import LandingPage from "./pages/LandingPage";
import Login from "./pages/Login";
import Signup from "./pages/Signup";  
import Home from "./pages/Home";
import Profile from "./pages/Profile";
import SavedPosts from "./pages/SavedPosts";
import SearchPage from './pages/SearchPage';
import ChatPage from './pages/ChatPage';

function App() {
  return (
    <Router>
      <Routes>
        <Route path="/" element={<LandingPage />} />
        <Route path="/login" element={<Login />} />
        <Route path="/signup" element={<Signup />} />
        <Route path="/home" element={<Home />} />
        <Route path="/profile" element={<Profile />} />
        <Route path="/profile/:userId" element={<Profile />} />
        <Route path="/saved-posts" element={<SavedPosts />} />
        <Route path="/search" element={<SearchPage />} />
        <Route path="/chat/:userId/:username" element={<ChatPage />} />
      </Routes>
    </Router>
  );
}

export default App;
