import React, { useState } from "react";
import axios from "axios";
import LikedUsersModal from "../components/LikedUsersModal";
import "bootstrap/dist/css/bootstrap.min.css";

const SearchPage = () => {
  const [searchTerm, setSearchTerm] = useState("");
  const [searchResults, setSearchResults] = useState([]);
  const [showResultsModal, setShowResultsModal] = useState(false);

  const handleKeyDown = async (e) => {
    if (e.key === "Enter") {
      try {
        const token = localStorage.getItem("jwtToken");
        const response = await axios.post(
          "http://localhost:8080/v1/users/search",
          {
            name: searchTerm,
          },
          {
            headers: {
              Authorization: `Bearer ${token}`,
              "Content-Type": "application/json",
            },
          }
        );

        setSearchResults(response.data.data || []);
        setShowResultsModal(true);
      } catch (error) {
        console.error("Error fetching search results:", error);
      }
    }
  };

  return (
    <div style={{ height: "100vh", display: "flex", flexDirection: "column" }}>
      {/* Top 10%: Search Box */}
      <div
        className="d-flex justify-content-center align-items-center"
        style={{ height: "10vh" }}
      >
        <input
          type="text"
          placeholder="Search users..."
          className="form-control"
          style={{ maxWidth: "500px", width: "100%" }}
          value={searchTerm}
          onChange={(e) => setSearchTerm(e.target.value)}
          onKeyDown={handleKeyDown}
        />
      </div>

      {/* Bottom 90%: Reserved for future features */}
      <div style={{ flex: "1", padding: "10px" }}>
        {/* Content to be added later */}
      </div>

      {/* LikedUsersModal to display search results */}
      <LikedUsersModal
        show={showResultsModal}
        likedUsers={searchResults}
        onClose={() => setShowResultsModal(false)}
        title="Search Results"
        emptytitle="users"
      />
    </div>
  );
};

export default SearchPage;
