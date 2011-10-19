package org.tobarsegais.webapp;

import org.tobarsegais.webapp.data.Extension;
import org.tobarsegais.webapp.data.Plugin;
import org.tobarsegais.webapp.data.Toc;

import javax.servlet.ServletContext;
import javax.servlet.ServletContextEvent;
import javax.servlet.ServletContextListener;
import javax.xml.stream.XMLStreamException;
import java.io.IOException;
import java.net.HttpURLConnection;
import java.net.JarURLConnection;
import java.net.MalformedURLException;
import java.net.URL;
import java.net.URLConnection;
import java.util.Collections;
import java.util.LinkedHashMap;
import java.util.Map;
import java.util.Set;
import java.util.jar.JarEntry;
import java.util.jar.JarFile;

/**
 * Loads all the bundles.
 */
public class ServletContextListenerImpl implements ServletContextListener {
    public void contextInitialized(ServletContextEvent sce) {
        ServletContext application = sce.getServletContext();
        Map<String, Toc> contents = new LinkedHashMap<String, Toc>();
        for (String path : (Set<String>) application.getResourcePaths("/WEB-INF/bundles")) {
            if (path.endsWith(".jar")) {
                String key = path.substring("/WEB-INF/bundles/".length(), path.lastIndexOf(".jar"));
                application.log("Parsing " + path);
                URLConnection connection = null;
                try {
                    URL url = new URL("jar:" + application.getResource(path) + "!/");
                    connection = url.openConnection();
                    if (!(connection instanceof JarURLConnection)) {
                        application.log(path + " is not a jar file, ignoring");
                        continue;
                    }
                    JarURLConnection jarConnection = (JarURLConnection) connection;
                    JarFile jarFile = jarConnection.getJarFile();
                    JarEntry pluginEntry = jarFile.getJarEntry("plugin.xml");
                    if (pluginEntry == null) {
                        application.log(path + " does not contain a plugin.xml file, ignoring");
                        continue;
                    }
                    Plugin plugin = Plugin.read(jarFile.getInputStream(pluginEntry));
                    Extension toc = plugin.getExtension("org.eclipse.help.toc");
                    if (toc == null || toc.getFile("toc") == null) {
                        application.log(path + " does not contain a 'org.eclipse.help.toc' extension, ignoring");
                        continue;
                    }
                    JarEntry tocEntry = jarFile.getJarEntry(toc.getFile("toc"));
                    if (tocEntry == null) {
                        application.log(path + " is missing the referenced toc: " + toc.getFile("toc") + ", ignoring");
                        continue;
                    }
                    contents.put(key, Toc.read(jarFile.getInputStream(tocEntry)));
                    application.log(path + " successfully parsed and added as " + key);
                } catch (XMLStreamException e) {
                    application.log("Could not parse " + path + " due to " + e.getMessage(), e);
                } catch (MalformedURLException e) {
                    application.log("Could not parse " + path + " due to " + e.getMessage(), e);
                } catch (IOException e) {
                    application.log("Could not parse " + path + " due to " + e.getMessage(), e);
                } finally {
                    if (connection instanceof HttpURLConnection) {
                        // should never be the case, but we should try to be sure
                        ((HttpURLConnection) connection).disconnect();
                    }
                }
            }
        }
        application.setAttribute("toc", Collections.unmodifiableMap(contents));
    }

    public void contextDestroyed(ServletContextEvent sce) {
        //To change body of implemented methods use File | Settings | File Templates.
    }
}
