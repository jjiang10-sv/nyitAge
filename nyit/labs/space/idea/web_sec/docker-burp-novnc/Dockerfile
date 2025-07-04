# Burp Suite with noVNC (Browser-based VNC)
FROM ubuntu:20.04

# Avoid interactive prompts during package installation
ENV DEBIAN_FRONTEND=noninteractive

# Add OpenJDK 17 repository and install packages
RUN apt-get update && apt-get install -y \
    software-properties-common \
    && add-apt-repository ppa:openjdk-r/ppa \
    && apt-get update && apt-get install -y \
    openjdk-17-jdk \
    wget \
    curl \
    unzip \
    xvfb \
    x11vnc \
    fluxbox \
    dbus-x11 \
    supervisor \
    novnc \
    websockify \
    net-tools \
    netcat \
    x11-utils \
    xauth \
    && rm -rf /var/lib/apt/lists/*

# Set JAVA_HOME
ENV JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64
ENV PATH=$JAVA_HOME/bin:$PATH

# Create burp user
RUN useradd -m -s /bin/bash burp

# Set working directory
WORKDIR /home/burp

# Download Burp Suite Community Edition (Java 17 compatible version)
RUN wget -O burpsuite_community.jar \
    "https://portswigger-cdn.net/burp/releases/download?product=community&version=2023.10.3.7&type=Jar"

# Create necessary directories
RUN mkdir -p /home/burp/.BurpSuite \
    && mkdir -p /tmp/certificates \
    && mkdir -p /var/log/supervisor \
    && mkdir -p /tmp/.X11-unix \
    && chmod 1777 /tmp/.X11-unix

# Set up noVNC
RUN ln -s /usr/share/novnc/vnc.html /usr/share/novnc/index.html

# Copy configuration files
COPY config/ /home/burp/.BurpSuite/
COPY supervisord.conf /etc/supervisor/conf.d/supervisord.conf
COPY scripts/ /home/burp/

# Make scripts executable
RUN chmod +x /home/burp/*.sh

# Change ownership
RUN chown -R burp:burp /home/burp

# Environment variables
ENV DISPLAY=:0 \
    VNC_RESOLUTION=1280x720 \
    VNC_PASSWORD=burp123

# Expose ports
EXPOSE 8082 5900 6080

# Start supervisor
CMD ["/usr/bin/supervisord", "-c", "/etc/supervisor/conf.d/supervisord.conf"] 