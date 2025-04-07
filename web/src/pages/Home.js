import React, { useState, useEffect } from "react";
import { useNavigate } from "react-router-dom";
import axios from "axios";
import { FiUser, FiLogOut } from "react-icons/fi";
import PostCard from "../components/PostCard";
import "bootstrap/dist/css/bootstrap.min.css";

const Home = () => {
  const [posts, setPosts] = useState([]);
  const navigate = useNavigate();

  useEffect(() => {
    const fetchPosts = async () => {
      try {
        const token = localStorage.getItem("jwtToken");
        const response = await axios.get("http://localhost:8080/v1/users/feed", {
          headers: { Authorization: `Bearer ${token}` },
        });

        setPosts(response.data.data || []);
      } catch (error) {
        console.error("Error fetching posts:", error);
      }
    };

    fetchPosts();
  }, []);

  const handleLogout = async () => {
    try {
      const token = localStorage.getItem("jwtToken");
      await axios.post("http://localhost:8080/v1/users/logout", {}, {
        headers: { Authorization: `Bearer ${token}` },
      });

      localStorage.removeItem("jwtToken");
      navigate("/login");
    } catch (error) {
      console.error("Logout failed:", error);
    }
  };

  return (
    <div className="container mt-4">
      <div className="d-flex justify-content-between align-items-center">
        <FiUser
          size={24}
          style={{ cursor: "pointer" }}
          className="text-primary"
          onClick={() => navigate("/profile")}
        />

        <FiLogOut
          size={24}
          style={{ cursor: "pointer" }}
          className="text-danger"
          onClick={handleLogout}
        />
      </div>

      <h2 className="text-center mb-4">User Feed</h2>

      <div className="row">
        {posts.map((post) => (
          <div className="col-md-4" key={post.id}>
            <PostCard post={post} />
          </div>
        ))}
      </div>
    </div>
  );
};

export default Home;
