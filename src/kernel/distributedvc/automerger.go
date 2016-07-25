package distributedvc

// components for invocating auto-merge automatically. Should be launched in a seperate goroutine
// and schedule mewrging work periodically.

import (
    "sync"
    "errors"
    conf "definition/configinfo"
    gspdi "intranet/gossipd/interactive"
    gsp "intranet/gossip"
    "outapi"
    . "outapi"
    . "logger"
    "strconv"
    "strings"
    "time"
)


type taskNode struct {
    prev *taskNode
    next *taskNode
    taskID string
}
// This class is used for maintaining marging task order for a lot of merging requests
// from FDs
type MergingScheduler struct {
    lock *sync.RWMutex

    // existMap, if map[key]==true, the key has not been checked-out or has been and
    // no new identical task is checked in during the working time
    // if map[key]==false, an identical task has been checked-in, so when commiting,
    // another inspection is needed.
    existMap map[string]bool
    taskQueue chan string
}

func NewMergingScheduler() *MergingScheduler {
    var ret=&MergingScheduler {
        lock: &sync.RWMutex{},
        existMap: make(map[string]bool),
        taskQueue: make(chan string, conf.AUTO_MERGER_TASK_QUEUE_CAPACITY),
    }

    return ret
}

var QUEUE_CAPACITY_REACHED=errors.New("The task queue is filled up.")
// if err!=nil, the Scheduler simply reject its task check-in request.
func (this *MergingScheduler)CheckInATask(filename string, io Outapi) error {
    this.lock.Lock()
    defer this.lock.Unlock()

    var id=genID_static(filename, io)

    if _, ok:=this.existMap[id]; ok {
        // the task has existed in the queue. DO NOT NEED to check in it again.
        this.existMap[id]=false
        return nil
    }


    if len(this.taskQueue)>=conf.AUTO_MERGER_TASK_QUEUE_CAPACITY {
        return QUEUE_CAPACITY_REACHED
    }
    this.taskQueue<-id

    this.existMap[id]=true

    return nil
}
func (this *MergingScheduler)CheckInATaskX(id string) error {
    this.lock.Lock()
    defer this.lock.Unlock()

    if _, ok:=this.existMap[id]; ok {
        // the task has existed in the queue. DO NOT NEED to check in it again.
        this.existMap[id]=false
        return nil
    }


    if len(this.taskQueue)>=conf.AUTO_MERGER_TASK_QUEUE_CAPACITY {
        return QUEUE_CAPACITY_REACHED
    }
    this.taskQueue<-id

    this.existMap[id]=true

    return nil
}

var NO_TASK_AVAILABLE=errors.New("No task is available in the queue.")
// if no available task, it blocks
func (this *MergingScheduler)ChechOutATask() string {
    return <-this.taskQueue
}

// if returns==false, it is needed to inspect the task again.
// otherwise, the task is successfully removed.
func (this *MergingScheduler)FinishTask(taskID string) bool {
    this.lock.Lock()
    defer this.lock.Unlock()

    if val, ok:=this.existMap[taskID]; !ok {
        panic("UNEXPECTED LOGICAL FLOW!")
    } else {
        if val {
            delete(this.existMap, taskID)
            return true
        } else {
            this.existMap[taskID]=true
            return false
        }
    }
}

// =============================================================================
// =============================================================================
// =============================================================================

type MergingSupervisor struct {
    lock *sync.RWMutex

    workersAlive int
    scheduler *MergingScheduler
    deamoned bool

    mpLock *sync.RWMutex
    gossipedMap map[string]*gspdi.GossipEntry
    mapIDCount int
}

var MergeManager=&MergingSupervisor {
    lock: &sync.RWMutex{},
    workersAlive: 0,
    scheduler: NewMergingScheduler(),

    mpLock: &sync.RWMutex{},
    gossipedMap: make(map[string]*gspdi.GossipEntry),
    mapIDCount: 0,
}

func (this *MergingSupervisor)Reveal_workersAlive() int {
    this.lock.RLock()
    defer this.lock.RUnlock()

    return this.workersAlive
}
const (
    REVEALED_TASK_IN_WORK=1
)
func (this *MergingSupervisor)Reveal_taskInfo() map[string]int {
    this.scheduler.lock.RLock()
    defer this.scheduler.lock.RUnlock()

    var ret=make(map[string]int)
    for k, _:=range this.scheduler.existMap {
        ret[k]=REVEALED_TASK_IN_WORK
    }

    return ret
}

const GOSSIP_TASK_PREFIX="$GOSSIPED_TASK$-#"
func (this *MergingSupervisor)SubmitGossipingTask(taskInfo *gspdi.GossipEntry) error {
    this.mpLock.Lock()
    this.mapIDCount++
    var cname=GOSSIP_TASK_PREFIX+strconv.Itoa(this.mapIDCount)
    this.gossipedMap[cname]=taskInfo
    this.mpLock.Unlock()
    if err:=this.scheduler.CheckInATaskX(cname); err!=nil {
        Secretary.Warn("distributedvc::MergingSupervisor.SubmitTask", "Failed to checkin task <"+cname+">: "+err.Error())
        return err
    }

    return nil
}
func (this *MergingSupervisor)SubmitTask(filename string, io Outapi) error {
    //Insider.Log("MergingSupervisor.SubmitTask()", "Start")
    if err:=this.scheduler.CheckInATask(filename, io); err!=nil {
        Secretary.Warn("distributedvc::MergingSupervisor.SubmitTask", "Failed to checkin task <"+filename+", "+io.GenerateUniqueID()+">: "+err.Error())
        return err
    }

    return nil
}

func (this *MergingSupervisor)__deprecated__reportDeath() {
    this.lock.Lock()
    defer this.lock.Unlock()
    this.workersAlive--
}

func (this *MergingSupervisor)__deprecated__spawnWorker() {
    this.lock.RLock()
    if this.workersAlive>=conf.MAX_MERGING_WORKER {
        this.lock.RUnlock()
        return
    }
    this.lock.RUnlock()

    this.lock.Lock()
    defer this.lock.Unlock()
    if this.workersAlive>=conf.MAX_MERGING_WORKER {
        return
    }
    this.workersAlive++
    go workerProcess(this, this.workersAlive)
}

// periodically spawn a worker to finish unadopted tasks
func (this *MergingSupervisor)Launch() {
    this.lock.Lock()
    defer this.lock.Unlock()
    for i:=0; i<conf.MAX_MERGING_WORKER; i++ {
        go workerProcess(this, i)
    }
    Secretary.Log("distributedvc::MergingSupervisor.Launch()", "Workers #1 to #"+strconv.Itoa(conf.MAX_MERGING_WORKER)+" have all been launched.")
    this.workersAlive+=conf.MAX_MERGING_WORKER
}

// =============================================================================

var worker_Sleep_Duration=time.Millisecond*time.Duration(conf.REST_INTERVAL_OF_WORKER_IN_MS)
func workerProcess(supervisor *MergingSupervisor, numbered int) {
    var myName="Merger worker #"+strconv.Itoa(numbered)
    // Secretary.Log(myName, "Worker is launched.")
    for {
        // loop until there is no task available
        var task=supervisor.scheduler.ChechOutATask()

        Secretary.Log(myName, "Got task:   "+task)
        if strings.HasPrefix(task, GOSSIP_TASK_PREFIX) {

            // a gossiped task
            supervisor.mpLock.RLock()
            var e=supervisor.gossipedMap[task]
            supervisor.mpLock.RUnlock()
            if e==nil {
                Secretary.Error(myName, "Logical error: fail to get a supposed-to-be gossiping task.")
                continue
            }

            if io:=outapi.DeSerializeID(e.OutAPI); io==nil {
                Secretary.Warn(myName, "Invalid Outapi DeSerializing: "+e.OutAPI)
                continue
            } else {
                if fd:=GetFD(e.Filename, io); fd==nil {
                    Secretary.Warn(myName, "Fail to get FD for "+e.Filename)
                    continue
                } else {
                    fd.GraspReader()
                    fd.ASYNCMergeWithNodeX(e, func(rse int) {
                        if rse==1 {
                            if err:=gsp.GlobalGossiper.PostGossip(e); err!=nil {
                                Secretary.Warn(myName, "Fail to post change gossiping to other nodes: "+err.Error())
                            }
                        } else if rse==2 {
                            // the file itself needs gossiping. wait for it to writeback and trigger gossiping
                            // DO NOTHING now
                        }
                    })
                    fd.WriteBack()
                    fd.ReleaseReader()
                    fd.Release()
                    Secretary.Log(myName, "Gossip processed: "+e.Filename+" @ "+e.OutAPI)
                }
            }
            if supervisor.scheduler.FinishTask(task) {
                supervisor.mpLock.Lock()
                delete(supervisor.gossipedMap, task)
                supervisor.mpLock.Unlock()
                Secretary.Log(myName, "Successfully accomplished task:    "+task)
                continue
            }

        } else {

            // a merging task
            var writeBackCount=0
            for {
                // loop until the task is removed from tasklist
                var thisFD=PeepFDX(task)
                if thisFD!=nil {
                    thisFD.GraspReader()
                    for {
                        // loop until there's nothing to merge for the fd
                        var merr=thisFD.MergeNext()
                        if merr!=nil {
                            if merr==NOTHING_TO_MERGE {
                                break
                            }
                            // ERROR when merge: Attentez: in such circumenstance,
                            // the patch may be on the way of submission
                            break
                        }
                        Secretary.Log(myName, "FD "+task+" has been merged once.")
                        writeBackCount++
                        if writeBackCount>=conf.AUTO_COMMIT_PER_INTRAMERGE {
                            writeBackCount=0
                            thisFD.WriteBack()
                            Secretary.Log(myName, "FD "+task+" has been written back once.")
                        }
                        time.Sleep(worker_Sleep_Duration)
                    }
                    thisFD.WriteBack()
                    Secretary.Log(myName, "FD "+task+" has been written back once.")

                    thisFD.ReleaseReader()
                    thisFD.Release()
                } else {
                    Secretary.Log(myName, "FD "+task+" is not in the fdPool. Abort.")
                }
                if supervisor.scheduler.FinishTask(task) {
                    Secretary.Log(myName, "Successfully accomplished task:    "+task)
                    break
                }
            }
        }
    }

    return
}
