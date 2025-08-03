// Navigation functionality
class Navigation {
  constructor() {
    console.log('Navigation constructor called');
    this.sidebar = document.getElementById('sidebar');
    this.menuToggle = document.getElementById('menuToggle');
    this.overlay = document.getElementById('sidebarOverlay');
    this.mainContent = document.getElementById('mainContent');
    this.isOpen = false;
    
    console.log('Elements found:', {
      sidebar: !!this.sidebar,
      menuToggle: !!this.menuToggle,
      overlay: !!this.overlay,
      mainContent: !!this.mainContent
    });
    
    this.init();
  }
  
  init() {
    console.log('Initializing navigation...');
    
    // Set active page based on current URL
    this.setActivePage();
    
    // Add event listeners
    if (this.menuToggle) {
      this.menuToggle.addEventListener('click', () => {
        console.log('Menu toggle clicked');
        this.toggleSidebar();
      });
    } else {
      console.error('Menu toggle button not found!');
    }
    
    if (this.overlay) {
      this.overlay.addEventListener('click', () => {
        console.log('Overlay clicked');
        this.closeSidebar();
      });
    }
    
    // Close sidebar on escape key
    document.addEventListener('keydown', (e) => {
      if (e.key === 'Escape' && this.isOpen) {
        console.log('Escape key pressed');
        this.closeSidebar();
      }
    });
    
    // Handle window resize
    window.addEventListener('resize', () => this.handleResize());
    
    // Initialize responsive behavior
    this.handleResize();
    
    console.log('Navigation initialization complete');
  }
  
  toggleSidebar() {
    console.log('Toggle sidebar, current state:', this.isOpen);
    if (this.isOpen) {
      this.closeSidebar();
    } else {
      this.openSidebar();
    }
  }
  
  openSidebar() {
    console.log('Opening sidebar');
    this.isOpen = true;
    if (this.sidebar) this.sidebar.classList.add('open');
    if (this.overlay) this.overlay.classList.add('open');
    if (this.mainContent) this.mainContent.classList.add('sidebar-open');
    if (this.menuToggle) this.menuToggle.innerHTML = '✕';
  }
  
  closeSidebar() {
    console.log('Closing sidebar');
    this.isOpen = false;
    if (this.sidebar) this.sidebar.classList.remove('open');
    if (this.overlay) this.overlay.classList.remove('open');
    if (this.mainContent) this.mainContent.classList.remove('sidebar-open');
    if (this.menuToggle) this.menuToggle.innerHTML = '☰';
  }
  
  handleResize() {
    const isMobile = window.innerWidth <= 768;
    console.log('Window resize detected, isMobile:', isMobile);
    
    if (isMobile) {
      // On mobile, sidebar should be hidden by default
      this.closeSidebar();
    } else {
      // On desktop, sidebar should be visible by default
      if (this.sidebar) this.sidebar.classList.add('open');
      if (this.mainContent) this.mainContent.classList.add('sidebar-open');
      if (this.menuToggle) this.menuToggle.innerHTML = '✕';
      this.isOpen = true;
    }
  }
  
  setActivePage() {
    const currentPath = window.location.pathname;
    console.log('Setting active page for path:', currentPath);
    const navItems = document.querySelectorAll('.nav-item');
    
    navItems.forEach(item => {
      const href = item.getAttribute('href');
      if (href === currentPath || 
          (currentPath === '/web/' && href === '/web/') ||
          (currentPath === '/web' && href === '/web/') ||
          (currentPath === '/web/list/' && href === '/web/list/') ||
          (currentPath === '/web/list' && href === '/web/list/') ||
          (currentPath === '/web/dashboard/' && href === '/web/dashboard/') ||
          (currentPath === '/web/dashboard' && href === '/web/dashboard/')) {
        item.classList.add('active');
        console.log('Set active:', href);
      } else {
        item.classList.remove('active');
      }
    });
  }
}

// Initialize navigation when DOM is loaded
document.addEventListener('DOMContentLoaded', () => {
  console.log('DOM Content Loaded - Initializing navigation...');
  new Navigation();
});

// Also initialize if DOM is already loaded
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', () => {
    console.log('DOM Loading - Initializing navigation...');
    new Navigation();
  });
} else {
  console.log('DOM Ready - Initializing navigation...');
  new Navigation();
}

// Export for potential use in other scripts
window.Navigation = Navigation; 