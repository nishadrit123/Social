// StoryViewerModal.js
import React, { useEffect, useState } from "react";
import "bootstrap/dist/css/bootstrap.min.css";

const StoryViewerModal = ({ show, onHide, stories }) => {
  const [currentIndex, setCurrentIndex] = useState(0);

  useEffect(() => {
    if (show && stories.length > 0) {
      const timer = setInterval(() => {
        setCurrentIndex((prevIndex) =>
          prevIndex + 1 < stories.length ? prevIndex + 1 : 0
        );
      }, 2000);

      return () => clearInterval(timer);
    }
  }, [show, stories]);

  if (!show || stories.length === 0) return null;

  const currentStory = stories[currentIndex];

  return (
    <div
      className="modal fade show d-block"
      tabIndex="-1"
      role="dialog"
      style={{ backgroundColor: "rgba(0,0,0,0.5)" }}
    >
      <div className="modal-dialog modal-dialog-centered" role="document">
        <div className="modal-content">
          <div className="modal-header">
            <h5 className="modal-title">{currentStory.title}</h5>
            <button
              type="button"
              className="btn-close"
              onClick={onHide}
              aria-label="Close"
            ></button>
          </div>
          <div className="modal-body">
            <p>{currentStory.content}</p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default StoryViewerModal;
