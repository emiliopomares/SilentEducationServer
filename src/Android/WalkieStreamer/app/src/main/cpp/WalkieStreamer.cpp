#include <jni.h>
#include <string>
#include <android/log.h>
#include <AndroidIO/SuperpoweredAndroidAudioIO.h>
#include <SuperpoweredSimple.h>
#include <SuperpoweredRecorder.h>
#include <malloc.h>
#include <arpa/inet.h>
#include <netinet/in.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/types.h>
#include <sys/socket.h>
#include <unistd.h>

#define MAXBUFLEN 8192
#define NPACK 10

unsigned int buflen;

JNIEnv *mainActivityEnv;
jobject mainActivityObj;
jmethodID updateMax;
jclass cls;


void diep(const char *s)
{
    perror(s);
    exit(1);
}
#define SRV_IP "192.168.1.43"
#define PORT 8080

struct sockaddr_in si_other;
int s, i, slen;
char buf[MAXBUFLEN];

#define BYTES_PER_SAMPLE 2

extern "C" JNIEXPORT void JNICALL
Java_com_silenteducation_streamer_MainActivity_initUDP(
        JNIEnv* env,
        jobject /* this */,
        jint bufferLength) {
    buflen = (unsigned int)bufferLength * BYTES_PER_SAMPLE;
    slen = sizeof(si_other);
    for(int i = 0; i < MAXBUFLEN; ++i) {
        buf[i] = 0;
    }
}

extern "C" JNIEXPORT jstring JNICALL
Java_com_silenteducation_streamer_MainActivity_connect(
        JNIEnv* env,
        jobject obj/* this */,
        jstring host,
        jint port) {

    mainActivityEnv = env;
    mainActivityObj = obj;
    cls = env->GetObjectClass(obj);//, mainActivityObj)
    updateMax = env->GetMethodID(cls, "updateMaxValue", "(I)V");

    const char *utfhost = env->GetStringUTFChars(host, 0);

    if((s=socket(AF_INET, SOCK_DGRAM, IPPROTO_UDP))==-1) {
        std::string response = "could not create socket";
        return env->NewStringUTF(response.c_str());
    }

    memset((char *)&si_other, 0, sizeof(si_other));
    si_other.sin_family = AF_INET;
    si_other.sin_port = htons((unsigned int)port);
    if(inet_aton("192.168.1.43", &si_other.sin_addr)==0) {
        std::string response = "could not parse address";
        return env->NewStringUTF(response.c_str());
    }
    std::string response = "connect OK";
    return env->NewStringUTF(response.c_str());
}

extern "C" JNIEXPORT void JNICALL
Java_com_silenteducation_streamer_MainActivity_sendPacket(
        JNIEnv* env,
        jobject /* this */) {
    for(int i = 0; i < NPACK; ++i) {
        if (sendto(s, buf, buflen, 0, (struct sockaddr *) &si_other, slen) == -1)
            diep("sendto()i failed");
    }

}

extern "C" JNIEXPORT void JNICALL
Java_com_silenteducation_streamer_MainActivity_cleanUp(
        JNIEnv* env,
        jobject /* this */) {
    close(s);
}


#define log_write __android_log_write

static SuperpoweredAndroidAudioIO *audioIO;
static SuperpoweredRecorder *recorder;
float *floatBuffer;

const int numberOfChannels = 2;

// This is called periodically by the audio engine.
static bool audioProcessing (
        void * __unused clientdata, // custom pointer
        short int *audio,           // buffer of interleaved samples
        int numberOfFrames,         // number of frames to process
        int __unused samplerate     // sampling rate
) {
    //SuperpoweredShortIntToFloat(audio, floatBuffer, (unsigned int)numberOfFrames);
    //recorder->process(floatBuffer, (unsigned int)numberOfFrames);
    sendto(s, (unsigned char *)audio, buflen * numberOfChannels, 0, (struct sockaddr *) &si_other, slen);
    return false;
}

// This is called after the recorder closed the WAV file.
static void recorderStopped (void * __unused clientdata) {
    log_write(ANDROID_LOG_DEBUG, "RecorderExample", "Finished recording.");
    delete recorder;
}

// StartAudio - Start audio engine.
extern "C" JNIEXPORT void
Java_com_silenteducation_streamer_MainActivity_StartAudio (
        JNIEnv *env,
        jobject  __unused obj,
        jint samplerate,
        jint buffersize,
        jstring tempPath,       // path to a temporary file
        jstring destPath        // path to the destination file
) {

    // Get path strings.
    const char *temp = env->GetStringUTFChars(tempPath, 0);
    const char *dest = env->GetStringUTFChars(destPath, 0);

    // Initialize the recorder with a temporary file path.
    recorder = new SuperpoweredRecorder (
            temp,               // The full filesystem path of a temporarily file.
            (unsigned int)samplerate,   // Sampling rate.
            1,                  // The minimum length of a recording (in seconds).
            2,                  // The number of channels.
            false,              // applyFade (fade in/out at the beginning / end of the recording)
            recorderStopped,    // Called when the recorder finishes writing after stop().
            NULL                // A custom pointer your callback receives (clientData).
    );

    // Start the recorder with the destination file path.
    recorder->start(dest);

    // Release path strings.
    env->ReleaseStringUTFChars(tempPath, temp);
    env->ReleaseStringUTFChars(destPath, dest);

    // Initialize float audio buffer.
    floatBuffer = (float *)malloc(sizeof(float) * 2 * buffersize);

    // Initialize audio engine with audio callback function.
    audioIO = new SuperpoweredAndroidAudioIO (
            samplerate,                     // sampling rate
            buffersize,                     // buffer size
            true,                           // enableInput
            false,                          // enableOutput
            audioProcessing,                // process callback function
            NULL                            // clientData
    );

    //mainActivityEnv->
    //mainActivityEnv->CallVoidMethod(mainActivityObj, updateMax, (int)76);

    sendto(s, buf, buflen, 0, (struct sockaddr *) &si_other, slen);
}

// StopAudio - Stop audio engine and free audio buffer.
extern "C" JNIEXPORT void
Java_com_silenteducation_streamer_MainActivity_StopAudio (
        JNIEnv * __unused env,
        jobject __unused obj
) {
    recorder->stop();
    delete audioIO;
    free(floatBuffer);
}

// onBackground - Put audio processing to sleep.
extern "C" JNIEXPORT void
Java_com_silenteducation_streamer_MainActivity_onBackground (
        JNIEnv * __unused env,
        jobject __unused obj
) {
    audioIO->onBackground();
}

// onForeground - Resume audio processing.
extern "C" JNIEXPORT void
Java_com_silenteducation_streamer_MainActivity_onForeground (
        JNIEnv * __unused env,
        jobject __unused obj
) {
    audioIO->onForeground();
}
