package org.tobarsegais.webapp;

import org.apache.commons.io.IOUtils;

import javax.servlet.ServletException;
import javax.servlet.http.HttpServlet;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.JarURLConnection;
import java.net.URL;
import java.net.URLConnection;
import java.util.jar.JarEntry;
import java.util.jar.JarFile;

/**
 * Created by IntelliJ IDEA.
 * User: stephenc
 * Date: 18/10/2011
 * Time: 09:35
 * To change this template use File | Settings | File Templates.
 */
public class ContentServlet extends HttpServlet {
    @Override
    protected void doGet(HttpServletRequest req, HttpServletResponse resp) throws ServletException, IOException {
        String path = req.getPathInfo();
        for (int index = path.indexOf('/'); index != -1; index = path.indexOf('/', index + 1)) {
            URL resource = getServletContext().getResource("/WEB-INF/bundles" + path.substring(0, index) + ".jar");
            if (resource == null) {
                continue;
            }
            URL jarResource = new URL("jar:" + resource + "!/");
            URLConnection connection = jarResource.openConnection();
            if (!(connection instanceof JarURLConnection)) {
                continue;
            }
            JarURLConnection jarConnection = (JarURLConnection) connection;
            JarFile jarFile = jarConnection.getJarFile();

            int endOfFileName = path.indexOf('#', index);
            endOfFileName = endOfFileName == -1 ? path.length() : endOfFileName;
            String fileName = path.substring(index+1, endOfFileName);
            JarEntry jarEntry = jarFile.getJarEntry(fileName);
            if (jarEntry == null) {
                continue;
            }
            InputStream in = null;
            OutputStream out = resp.getOutputStream();
            try {
                in = jarFile.getInputStream(jarEntry);
                IOUtils.copy(in, out);
            } finally {
                IOUtils.closeQuietly(in);
                out.close();
            }
            return;
        }
        resp.sendError(404);
    }
}
